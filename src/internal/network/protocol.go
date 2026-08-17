package network

import "encoding/json"

type MessageType string

const (
        MsgJoin        MessageType = "join"
        MsgWelcome     MessageType = "welcome"
        MsgPlayerLeft  MessageType = "player_left"
        MsgLobbyUpdate MessageType = "lobby_update"
        MsgStartGame   MessageType = "start_game"
        MsgGameStart   MessageType = "game_start"
        MsgInput       MessageType = "input"
        MsgStateUpdate MessageType = "state_update"
        MsgError       MessageType = "error"
)

type Envelope struct {
        Type    MessageType     `json:"type"`
        Payload json.RawMessage `json:"payload,omitempty"`
}

type JoinPayload struct {
        Name string `json:"name"`
        Role string `json:"role"` // "host" or "join", as chosen on the menu screen
}

type WelcomePayload struct {
        PlayerID int    `json:"playerId"`
        Faction  string `json:"faction"`
}

type LobbyPlayer struct {
        ID     int    `json:"id"`
        Name   string `json:"name"`
        IsHost bool   `json:"isHost"`
}

type LobbyUpdatePayload struct {
        Players []LobbyPlayer `json:"players"`
}

type InputPayload struct {
        // TODO: expand into real player commands (move, capture, build, ...)
        Command string          `json:"command"`
        Args    json.RawMessage `json:"args,omitempty"`
}

type StateUpdatePayload struct {
        Tick     uint64           `json:"tick"`
        Entities []EntitySnapshot `json:"entities"`
}

type EntitySnapshot struct {
        ID      string  `json:"id"`
        Type    string  `json:"type"`
        Faction string  `json:"faction"`
        X       float64 `json:"x"`
        Y       float64 `json:"y"`
        Z       float64 `json:"z"`
        Heading float64 `json:"heading"`
}
