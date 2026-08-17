package main

import (
        "encoding/json"
        "fmt"
        "log"
        "net"
        "net/http"

        "battle/src/internal/network"
        "battle/src/internal/sim"
)

// basePort is where the server starts looking for a free port; portTries is
// how many consecutive ports it will attempt before giving up.
const basePort = 8080
const portTries = 20

func main() {
        hub := network.NewHub()
        world := sim.NewWorld()

        go handleInbound(hub, world)
        go sim.Run(world, func(w *sim.World) {
                broadcastState(hub, w)
        })

        mux := http.NewServeMux()
        mux.HandleFunc("/ws", hub.ServeWS)
        mux.Handle("/", http.FileServer(http.Dir("src/web")))

        listener, addr := listenOnAvailablePort(basePort, portTries)
        log.Printf("host server listening on %s", addr)
        log.Fatal(http.Serve(listener, mux))
}

// listenOnAvailablePort tries basePort, basePort+1, ... until one binds
// successfully, so a second local instance (or anything else already on
// 8080) doesn't just crash the server.
func listenOnAvailablePort(basePort, tries int) (net.Listener, string) {
        for port := basePort; port < basePort+tries; port++ {
                addr := fmt.Sprintf(":%d", port)
                ln, err := net.Listen("tcp", addr)
                if err == nil {
                        return ln, addr
                }
                log.Printf("port %d unavailable (%v), trying next", port, err)
        }
        log.Fatalf("no available port found in range %d-%d", basePort, basePort+tries-1)
        return nil, ""
}

func handleInbound(hub *network.Hub, world *sim.World) {
        for msg := range hub.Inbound {
                switch msg.Envelope.Type {
                case network.MsgJoin:
                        var payload network.JoinPayload
                        _ = json.Unmarshal(msg.Envelope.Payload, &payload)
                        hub.SetPlayerInfo(msg.ClientID, payload.Name, payload.Role == "host")
                        log.Printf("player %d joined as %q (role=%s)", msg.ClientID, payload.Name, payload.Role)
                        // TODO: assign faction, spawn starting units, send MsgWelcome
                        broadcastLobby(hub)
                case network.MsgPlayerLeft:
                        broadcastLobby(hub)
                case network.MsgStartGame:
                        if !hub.IsHost(msg.ClientID) {
                                log.Printf("player %d tried to start the game but is not host", msg.ClientID)
                                continue
                        }
                        log.Printf("player %d started the game", msg.ClientID)
                        hub.Broadcast(network.Envelope{Type: network.MsgGameStart})
                case network.MsgInput:
                        // TODO: validate and apply command to world
                default:
                        log.Printf("unhandled message type %q from player %d", msg.Envelope.Type, msg.ClientID)
                }
        }
        _ = world
}

func broadcastLobby(hub *network.Hub) {
        payload, err := json.Marshal(network.LobbyUpdatePayload{Players: hub.Roster()})
        if err != nil {
                log.Printf("marshal lobby: %v", err)
                return
        }
        hub.Broadcast(network.Envelope{Type: network.MsgLobbyUpdate, Payload: payload})
}

func broadcastState(hub *network.Hub, w *sim.World) {
        snapshots := make([]network.EntitySnapshot, 0, len(w.Entities))
        for _, e := range w.Entities {
                snapshots = append(snapshots, network.EntitySnapshot{
                        ID:      e.ID,
                        Type:    string(e.Kind),
                        Faction: e.Faction, 
                        X:       e.Pos.X,
                        Y:       e.Pos.Y,
                        Z:       e.Pos.Z,
                        Heading: e.Heading,
                })
        }
        payload, err := json.Marshal(network.StateUpdatePayload{Tick: w.Tick, Entities: snapshots})
        if err != nil {
                log.Printf("marshal state: %v", err)
                return
        }
        hub.Broadcast(network.Envelope{Type: network.MsgStateUpdate, Payload: payload})
}
