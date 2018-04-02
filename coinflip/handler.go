package coinflip

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
	"github.com/mtdx/ns-ga/common"
	"github.com/mtdx/ns-ga/logger"
	"github.com/mtdx/ns-ga/validator"
)

// GetGamesHandler GET /coinflip
func GetGamesHandler(w http.ResponseWriter, r *http.Request) {
	dbconn := r.Context().Value("DBCONN").(*sql.DB)
	gamesresp, err := getOpenGames(dbconn)
	common.ValidateRenderResults(w, r, gamesresp, err)
}

// GetTopPlayersHandler GET /coinflip-top-players
func GetTopPlayersHandler(w http.ResponseWriter, r *http.Request) {
	dbconn := r.Context().Value("DBCONN").(*sql.DB)
	playersresp, err := getTopPlayers(dbconn)
	common.ValidateRenderResults(w, r, playersresp, err)
}

// GetGamesHistoryHandler GET /coinflip-history
func GetGamesHistoryHandler(w http.ResponseWriter, r *http.Request) {
	dbconn := r.Context().Value("DBCONN").(*sql.DB)
	_, claims, _ := jwtauth.FromContext(r.Context())
	gamesresp, err := getGamesHistory(dbconn, fmt.Sprintf("%v", claims["steam_id"]))
	common.ValidateRenderResults(w, r, gamesresp, err)
}

// WebsocketsHandler GET /coinflip
func WebsocketsHandler(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	wsconn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error.Println(err.Error())
		return
	}
	defer wsconn.Close()

	clients[wsconn] = true
	for {
		var msg WsInMessage
		err := wsconn.ReadJSON(&msg)
		if err != nil {
			delete(clients, wsconn)
			break
		}
	}
}

// BroadcastWs ...
func BroadcastWs() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// GameCreateHandler POST /coinflip
func GameCreateHandler(w http.ResponseWriter, r *http.Request) {
	dbconn := r.Context().Value("DBCONN").(*sql.DB)
	_, claims, _ := jwtauth.FromContext(r.Context())

	game := &GameRequest{}
	if err := render.Bind(r, game); err != nil {
		render.Render(w, r, common.ErrInvalidRequest(err))
		return
	}
	if err := validator.Validate(game); err != nil {
		render.Render(w, r, common.ErrInvalidRequest(err))
		return
	}

	gameopen, err := createGame(dbconn, game, claims["steam_id"])
	if err != nil || gameopen == nil {
		if err == nil {
			err = errors.New("Can't return Coinflip Game")
		}
		render.Render(w, r, common.ErrInternalServer(err))
		return
	}

	broadcast <- gameopen
	render.Render(w, r, common.SuccessCreatedResponse("Game Created"))
}

// JoinGameHandler PUT /coinflip/{gameId}
func JoinGameHandler(w http.ResponseWriter, r *http.Request) {
	dbconn := r.Context().Value("DBCONN").(*sql.DB)
	_, claims, _ := jwtauth.FromContext(r.Context())

	gamecompleted, err := joinGame(dbconn, fmt.Sprintf("%v", claims["steam_id"]), chi.URLParam(r, "gameID")); 
	if err != nil {
		render.Render(w, r, common.ErrInternalServer(err))
		return
	}

	broadcast <- gamecompleted
	render.Render(w, r, common.SuccessCreatedResponse("Game Joined"))
}

// DeleteGameHandler DELETE /coinflip/{gameId}
func DeleteGameHandler(w http.ResponseWriter, r *http.Request) {
	dbconn := r.Context().Value("DBCONN").(*sql.DB)
	_, claims, _ := jwtauth.FromContext(r.Context())

	gamedeleted, err := deleteGame(dbconn, fmt.Sprintf("%v", claims["steam_id"]), chi.URLParam(r, "gameID")); 
	if err != nil {
		render.Render(w, r, common.ErrInternalServer(err))
		return
	}

	broadcast <- gamedeleted
	render.Render(w, r, common.SuccessCreatedResponse("Game Deleted"))
}