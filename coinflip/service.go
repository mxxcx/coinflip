package coinflip

import (
	"database/sql"
	"errors"
	"fmt"
	"math"

	"github.com/mtdx/ns-ga/common"

	"github.com/go-chi/render"
	"github.com/mtdx/ns-ga/cryptorand"
	"github.com/mtdx/ns-ga/validator"
)

func createGame(dbconn *sql.DB, game *GameRequest, player1Id interface{}) (*GameOpen, error) {
	salt, err := cryptorand.GenerateRandomString(32)
	if err != nil {
		return nil, err
	}
	pick, ok := Picks[game.Pick]
	if !ok {
		return nil, errors.New("Invalid Pick")
	}
	result, err := cryptorand.Randnum(2)
	if err != nil {
		return nil, err
	}
	hash := cryptorand.Hash(fmt.Sprintf(`%v:%v`, salt, result+1))
	var lastInsertID uint64
	var displayName, avatar string
	err = common.Transact(dbconn, func(tx *sql.Tx) error {
		err := tx.QueryRow(`SELECT display_name, avatar_medium FROM steam_profile_cache WHERE steam_id = $1`, player1Id).Scan(&displayName, &avatar)
		if err != nil || err == sql.ErrNoRows {
			return err
		}

		var balance float64
		err = tx.QueryRow(`SELECT balance FROM users WHERE steam_id = $1 FOR UPDATE`, player1Id).Scan(&balance)
		if err != nil || err == sql.ErrNoRows {
			return err
		}
		if balance-math.Abs(float64(game.Amount)) < 0 {
			return errors.New("Not Enough Balance")
		}

		err = tx.QueryRow(`INSERT INTO coinflip_games (player1_id, pick, result, amount, hash, salt) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
			player1Id,
			pick,
			result+1,
			game.Amount,
			hash,
			salt,
		).Scan(&lastInsertID)
		if err != nil {
			return err
		}
		_, err = tx.Exec(`UPDATE users SET balance = balance - $1,
			last_balance_change_reason = json_build_object(
				'type', 'coinflip_gameopen',
				'id', $2::int
			)
		  WHERE steam_id = $3`,
			math.Abs(float64(game.Amount)),
			lastInsertID,
			player1Id,
		)

		return err
	})
	if err != nil {
		return nil, err
	}

	gameopen := &GameOpen{
		ID:          lastInsertID,
		Status:      `gameopen`,
		Player1Id:   fmt.Sprintf("%v", player1Id),
		DisplayName: displayName,
		Avatar:      avatar,
		Amount:      game.Amount,
		Hash:        hash,
	}
	if err := validator.Validate(gameopen); err != nil {
		return nil, err
	}
	return gameopen, err
}

func joinGame(dbconn *sql.DB, player2Id string, gameID string) (*GameCompleted, error) {
	game := &GameCompleted{}
	var pick, result int
	var payout, commision float64
	var displayName, avatar string
	err := common.Transact(dbconn, func(tx *sql.Tx) error {
		err := tx.QueryRow(`SELECT display_name, avatar_medium FROM steam_profile_cache WHERE steam_id = $1`, player2Id).Scan(&displayName, &avatar)
		if err != nil || err == sql.ErrNoRows {
			return err
		}

		err = tx.QueryRow("SELECT id, player1_id, player2_id, amount, pick, result, hash, salt FROM coinflip_games WHERE id = $1 FOR UPDATE", gameID).Scan(
			&game.ID,
			&game.Player1Id,
			&game.Player2Id,
			&game.Amount,
			&pick,
			&result,
			&game.Hash,
			&game.Salt,
		)
		if err != nil || err == sql.ErrNoRows {
			return err
		}
		if game.Player2Id.Valid == true || game.Player1Id == player2Id {
			return errors.New("Game Completed or Same PlayerId")
		}

		var balance float64
		err = tx.QueryRow(`SELECT balance FROM users WHERE steam_id = $1 FOR UPDATE`, player2Id).Scan(&balance)
		if err != nil || err == sql.ErrNoRows {
			return err
		}
		if balance-math.Abs(float64(game.Amount)) < 0 {
			return errors.New("Not Enough Balance")
		}

		_, err = tx.Exec(`UPDATE users SET balance = balance - $1,
			last_balance_change_reason = json_build_object(
				'type', 'coinflip_gamejoin',
				'id', $2::int
			)
		  WHERE steam_id = $3`,
			math.Abs(float64(game.Amount)),
			gameID,
			player2Id,
		)
		if err != nil {
			return err
		}

		payout, commision = calcPayout(game.Amount)
		var winingplayer string
		if pick == result {
			winingplayer = game.Player1Id
		} else {
			winingplayer = player2Id
		}
		_, err = tx.Exec(`UPDATE users SET balance = balance + $1,
			last_balance_change_reason = json_build_object(
				'type', 'coinflip_gamecompleted',
				'id', $2::int
			)
		  WHERE steam_id = $3`,
			payout,
			gameID,
			winingplayer,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec("UPDATE coinflip_games SET player2_id = $1, commision = $2, payout = $3 WHERE id = $4",
			player2Id,
			commision,
			payout,
			gameID,
		)

		return err
	})
	if err != nil {
		return nil, err
	}

	game.Player2Id = sql.NullString{String: player2Id, Valid: true}
	game.DisplayName = displayName
	game.Avatar = avatar
	game.Pick = PicksRev[pick]
	game.Result = PicksRev[result]
	game.Status = `gamecompleted`
	game.Payout = payout
	game.Commision = commision

	if err := validator.Validate(game); err != nil {
		return nil, err
	}
	return game, err
}

func getOpenGames(dbconn *sql.DB) ([]render.Renderer, error) {
	rows, err := dbconn.Query(`SELECT id, player1_id, display_name, avatar_medium, amount, hash FROM coinflip_games 
		LEFT JOIN steam_profile_cache ON player1_id = steam_id WHERE player2_id IS NULL ORDER BY id ASC OFFSET 0 LIMIT 40`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	gamesresp := []render.Renderer{}
	for rows.Next() {
		resp := &GameOpen{}
		if err := rows.Scan(
			&resp.ID,
			&resp.Player1Id,
			&resp.DisplayName,
			&resp.Avatar,
			&resp.Amount,
			&resp.Hash,
		); err != nil {
			return nil, err
		}
		resp.Status = `gameopen`
		gamesresp = append(gamesresp, resp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return gamesresp, nil
}

func getTopPlayers(dbconn *sql.DB) ([]render.Renderer, error) {
	rows, err := dbconn.Query(`SELECT Sum(amount) as amount, player1_id as playerId FROM coinflip_games
	WHERE pick = result GROUP BY playerId ORDER BY amount DESC LIMIT 2`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	playersresp := []render.Renderer{}
	rawplayersresp := []*TopPlayer{}
	rawplayersresp, err = appendTopPlayers(rows, rawplayersresp)
	if err != nil {
		return nil, err
	}

	rows2, err := dbconn.Query(`SELECT Sum(amount) as amount, player2_id as playerId FROM coinflip_games
	WHERE pick != result GROUP BY playerId ORDER BY amount DESC LIMIT 2`)
	if err != nil {
		return nil, err
	}
	defer rows2.Close()
	rawplayersresp, err = appendTopPlayers(rows2, rawplayersresp)
	if err != nil {
		return nil, err
	}

	len := len(rawplayersresp)
	for i := 0; i < len; i++ {
		for j := i + 1; j < len; j++ {
			if rawplayersresp[j].Amount > rawplayersresp[i].Amount {
				rawplayersresp[i], rawplayersresp[j] = rawplayersresp[j], rawplayersresp[i]
			}
		}
	}
	for i := 0; i < 2 && i < len; i++ {
		playersresp = append(playersresp, rawplayersresp[i])
	}
	return playersresp, nil
}

func getGamesHistory(dbconn *sql.DB, userID string) ([]render.Renderer, error) {
	rows, err := dbconn.Query(`SELECT id, player1_id, player2_id, display_name, avatar_medium, amount, pick, result, hash, salt FROM coinflip_games
		LEFT JOIN steam_profile_cache ON player2_id = steam_id WHERE player2_id IS NOT NULL AND (player1_id = $1 OR player2_id = $1) 
		AND amount != 0 AND commision != 0 ORDER BY id DESC OFFSET 0 LIMIT 60`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	gamesresp := []render.Renderer{}
	var pick, result int
	for rows.Next() {
		resp := &GameCompleted{}
		if err := rows.Scan(
			&resp.ID,
			&resp.Player1Id,
			&resp.Player2Id,
			&resp.DisplayName,
			&resp.Avatar,
			&resp.Amount,
			&pick,
			&result,
			&resp.Hash,
			&resp.Salt,
		); err != nil {
			return nil, err
		}
		payout, commission := calcPayout(resp.Amount)
		resp.Pick = PicksRev[pick]
		resp.Result = PicksRev[result]
		resp.Status = `gamecompleted`
		resp.Payout = payout
		resp.Commision = commission
		gamesresp = append(gamesresp, resp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return gamesresp, nil
}

func deleteGame(dbconn *sql.DB, player1Id string, gameID string) (*GameDeleted, error) {
	var id uint64
	err := common.Transact(dbconn, func(tx *sql.Tx) error {
		var player2Id sql.NullString
		var amount float64
		err := tx.QueryRow("SELECT id, player2_id, amount FROM coinflip_games WHERE id = $1 AND player1_id = $2 AND player2_id IS NULL FOR UPDATE",
			gameID,
			player1Id,
		).Scan(
			&id,
			&player2Id,
			&amount,
		)
		if err != nil || err == sql.ErrNoRows {
			return err
		}
		if player2Id.Valid == true {
			return errors.New("Game Completed")
		}

		_, err = tx.Exec("DELETE FROM coinflip_games WHERE id = $1 AND player1_id = $2 AND player2_id IS NULL",
			gameID,
			player1Id,
		)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`UPDATE users SET balance = balance + $1,
			last_balance_change_reason = json_build_object(
				'type', 'coinflip_gamedeleted',
				'id', $2::int
			)
		  WHERE steam_id = $3`,
			math.Abs(amount),
			gameID,
			player1Id,
		)

		return err
	})
	if err != nil {
		return nil, err
	}

	game := &GameDeleted{
		ID:        id,
		Status:    `gamedeleted`,
		Player1Id: player1Id,
	}

	if err := validator.Validate(game); err != nil {
		return nil, err
	}
	return game, err
}

func appendTopPlayers(rows *sql.Rows, playersresp []*TopPlayer) ([]*TopPlayer, error) {
	for rows.Next() {
		resp := &TopPlayer{}
		if err := rows.Scan(
			&resp.Amount,
			&resp.PlayerID,
		); err != nil {
			return nil, err
		}
		playersresp = append(playersresp, resp)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return playersresp, nil
}

func calcPayout(amount uint) (float64, float64) {
	commision := math.Abs(float64(amount*2)) * COMMISSION
	payout := math.Abs(float64(amount*2)) - commision
	return payout, commision
}
