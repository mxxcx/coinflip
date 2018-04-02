package coinflip

import "net/http"
import "database/sql"

// Picks ...
var Picks = map[string]int{
	"red":  1,
	"blue": 2,
}

// PicksRev ...
var PicksRev = map[int]string{
	1: "red",
	2: "blue",
}

// GameRequest ...
type GameRequest struct {
	Pick   string `json:"pick" validate:"nonzero,regexp=^(red|blue)$"`
	Amount uint   `json:"amount" validate:"min=1"`
}

// GameOpen ...
type GameOpen struct {
	ID          uint64 `json:"id" validate:"nonzero"`
	Status      string `json:"status" validate:"nonzero"`
	Player1Id   string `json:"player1_id" validate:"nonzero"`
	DisplayName string `json:"display_name" validate:"nonzero"`
	Avatar      string `json:"avatar" validate:"nonzero"`
	Amount      uint   `json:"amount" validate:"min=1"`
	Hash        string `json:"hash" validate:"nonzero"`
}

// GameCompleted ...
type GameCompleted struct {
	ID          uint64         `json:"id" validate:"nonzero"`
	Status      string         `json:"status" validate:"nonzero"`
	Player1Id   string         `json:"player1_id" validate:"nonzero"`
	Player2Id   sql.NullString `json:"player2_id"`
	DisplayName string         `json:"display_name" validate:"nonzero"`
	Avatar      string         `json:"avatar" validate:"nonzero"`
	Amount      uint           `json:"amount" validate:"min=1"`
	Pick        string         `json:"pick" validate:"nonzero,regexp=^(red|blue)$"`
	Result      string         `json:"result" validate:"nonzero,regexp=^(red|blue)$"`
	Hash        string         `json:"hash" validate:"nonzero"`
	Salt        string         `json:"salt" validate:"nonzero"`
	Payout      float64        `json:"payout" validate:"min=1"`
	Commision   float64        `json:"commision" validate:"nonzero"`
}

// GameDeleted ...
type GameDeleted struct {
	ID        uint64 `json:"id" validate:"nonzero"`
	Status    string `json:"status" validate:"nonzero"`
	Player1Id string `json:"player1_id" validate:"nonzero"`
}

// WsInMessage ...
type WsInMessage struct {
	Message string `json:"message"`
}

// TopPlayer ...
type TopPlayer struct {
	PlayerID string `json:"player_id" validate:"nonzero"`
	Amount   uint   `json:"amount" validate:"min=1"`
}

// Bind post-processing after a decoding.
func (g *GameRequest) Bind(r *http.Request) error {
	return nil
}

// Render pre-processing after a decoding.
func (g *GameOpen) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render pre-processing after a decoding.
func (g *TopPlayer) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

// Render pre-processing after a decoding.
func (g *GameCompleted) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
