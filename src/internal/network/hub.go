package network

import (
        "log"
        "net/http"
        "sort"
        "sync"

        "github.com/gorilla/websocket"
)

// MaxPlayers matches the "up to 4 players" design in README.txt.
const MaxPlayers = 4

var upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
        CheckOrigin:     func(r *http.Request) bool { return true }, // local LAN play only
}

type Client struct {
        ID     int
        Conn   *websocket.Conn
        Name   string
        IsHost bool
        send   chan Envelope
}

type ClientMessage struct {
        ClientID int
        Envelope Envelope
}

// Hub is the host's connection manager: the host process is the server,
// every player (including the host's own browser tab) connects to it
// as a WebSocket client.
type Hub struct {
        mu      sync.Mutex
        clients map[int]*Client
        nextID  int
        Inbound chan ClientMessage
}

func NewHub() *Hub {
        return &Hub{
                clients: make(map[int]*Client),
                Inbound: make(chan ClientMessage, 256),
        }
  }

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
        h.mu.Lock()
        if len(h.clients) >= MaxPlayers {
                h.mu.Unlock()
                http.Error(w, "game full", http.StatusServiceUnavailable)
                return
        }
        h.mu.Unlock()

        conn, err := upgrader.Upgrade(w, r, nil)
        if err != nil {
                log.Printf("upgrade failed: %v", err)
                return
        }

        h.mu.Lock()
        id := h.nextID
        h.nextID++
        client := &Client{ID: id, Conn: conn, send: make(chan Envelope, 64)}
        h.clients[id] = client
        h.mu.Unlock()

        log.Printf("player %d connected", id)

        go h.writePump(client)
        h.readPump(client)
}

func (h *Hub) readPump(c *Client) {
        defer h.disconnect(c)
        for {
                var env Envelope
                if err := c.Conn.ReadJSON(&env); err != nil {
                        return
                }
                h.Inbound <- ClientMessage{ClientID: c.ID, Envelope: env}
        }
}

func (h *Hub) writePump(c *Client) {
        defer c.Conn.Close()
        for env := range c.send {
                if err := c.Conn.WriteJSON(env); err != nil {
                        return
                }
        }
  }

func (h *Hub) disconnect(c *Client) {
        h.mu.Lock()
        delete(h.clients, c.ID)
        h.mu.Unlock()
        close(c.send)
        c.Conn.Close()
        log.Printf("player %d disconnected", c.ID)
        h.Inbound <- ClientMessage{ClientID: c.ID, Envelope: Envelope{Type: MsgPlayerLeft}}
}

// SetPlayerInfo records the name and declared host status a client sent in
// its join message. Host status is trusted as declared by the client, not
// arbitrated by the server (e.g. by connection order).
func (h *Hub) SetPlayerInfo(clientID int, name string, isHost bool) {
        h.mu.Lock()
        defer h.mu.Unlock()
        if c, ok := h.clients[clientID]; ok {
                c.Name = name
                c.IsHost = isHost
        }
}

// IsHost reports whether the given client declared itself host at join time.
func (h *Hub) IsHost(clientID int) bool {
        h.mu.Lock()
        defer h.mu.Unlock()
        c, ok := h.clients[clientID]
        return ok && c.IsHost
}

// Roster returns the current lobby, ordered by connection order.
func (h *Hub) Roster() []LobbyPlayer {
        h.mu.Lock()
        defer h.mu.Unlock()
        players := make([]LobbyPlayer, 0, len(h.clients))
        for _, c := range h.clients {
                players = append(players, LobbyPlayer{ID: c.ID, Name: c.Name, IsHost: c.IsHost})
        }
        sort.Slice(players, func(i, j int) bool { return players[i].ID < players[j].ID })
        return players
}

func (h *Hub) Broadcast(env Envelope) {
        h.mu.Lock()
        defer h.mu.Unlock()
        for _, c := range h.clients { select {
                case c.send <- env:
                default:
                        log.Printf("player %d send buffer full, dropping message", c.ID)
                }
        }
}

func (h *Hub) Send(clientID int, env Envelope) {
        h.mu.Lock()
        c, ok := h.clients[clientID]
        h.mu.Unlock()
        if !ok {
                return
        }
        select {
        case c.send <- env:
        default:
                log.Printf("player %d send buffer full, dropping message", clientID)
        }
}
