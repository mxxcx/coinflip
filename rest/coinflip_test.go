package rest

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mtdx/ns-ga/coinflip"
	"github.com/stretchr/testify/assert"
)

func coinflipGameTest(t *testing.T) {
	game1 := &coinflip.GameRequest{
		Pick:   `red`,
		Amount: 10,
	}
	game2 := &coinflip.GameRequest{
		Pick:   `blue`,
		Amount: 20,
	}
	game3 := &coinflip.GameRequest{
		Pick: `blue`,
	}
	game4 := &coinflip.GameRequest{
		Pick:   `redd`,
		Amount: 20,
	}
	game5 := &coinflip.GameRequest{
		Pick:   `blue`,
		Amount: 1001,
	}
	game6 := &coinflip.GameRequest{
		Pick:   `red`,
		Amount: 2001,
	}

	// Test Create
	callErrEndPoint(t, game3, "POST", "/api/coinflip", jwt1, "Amount: less than min")
	callErrEndPoint(t, game4, "POST", "/api/coinflip", jwt2, "Pick: regular expression mismatch")
	callOkEndPoint(t, game1, "POST", "/api/coinflip", jwt1, "Game Created")
	callOkEndPoint(t, game2, "POST", "/api/coinflip", jwt2, "Game Created")
	assert.Equal(t, float64(1000-game1.Amount), getBalance(testUserID1))
	assert.Equal(t, float64(2000-game2.Amount), getBalance(testUserID2))
	callErrEndPoint(t, game5, "POST", "/api/coinflip", jwt1, "Not Enough Balance")
	callErrEndPoint(t, game6, "POST", "/api/coinflip", jwt2, "Not Enough Balance")
	_, body = callEndPoint(t, ts, "GET", "/api/coinflip-history", nil, jwt1)
	assert.Equal(t, "[]", strings.TrimSpace(body), body)
	_, body = callEndPoint(t, ts, "GET", "/api/coinflip-history", nil, jwt2)
	assert.Equal(t, "[]", strings.TrimSpace(body), body)

	// Test Join
	_, body := callEndPoint(t, ts, "GET", "/api/coinflip", nil, "")
	gamesresp := make([]coinflip.GameOpen, 2)
	validateGamesList(t, body, gamesresp, 2)
	assert.Equal(t, game1.Amount, gamesresp[0].Amount, body)
	assert.Equal(t, testUserID2, gamesresp[1].Player1Id, body)

	assertAuth(t, ts, "PUT", fmt.Sprintf("/api/coinflip/%d", gamesresp[0].ID))
	assertAuth(t, ts, "PUT", fmt.Sprintf("/api/coinflip/%d", gamesresp[1].ID))

	callErrEndPoint(t, nil, "PUT", fmt.Sprintf("/api/coinflip/%d", gamesresp[0].ID), jwt1, "Game Completed or Same PlayerId")
	callErrEndPoint(t, nil, "PUT", fmt.Sprintf("/api/coinflip/%d", gamesresp[1].ID), jwt2, "Game Completed or Same PlayerId")

	callOkEndPoint(t, nil, "PUT", fmt.Sprintf("/api/coinflip/%d", gamesresp[0].ID), jwt2, "Game Joined")
	b2 := getBalance(testUserID2)
	b2win := float64(2000-game1.Amount-game2.Amount) + float64(game1.Amount*2) - (float64(game1.Amount*2) * coinflip.COMMISSION)
	assert.True(t, b2win == b2 || float64(2000-game1.Amount-game2.Amount) == b2)

	b1i := getBalance(testUserID1)
	callOkEndPoint(t, nil, "PUT", fmt.Sprintf("/api/coinflip/%d", gamesresp[1].ID), jwt1, "Game Joined")
	b1 := getBalance(testUserID1)
	b1win := b1i - float64(game2.Amount) + float64(game2.Amount*2) - (float64(game2.Amount*2) * coinflip.COMMISSION)
	assert.True(t, b1win == b1 || b1i-float64(game2.Amount) == b1)

	callErrEndPoint(t, nil, "PUT", fmt.Sprintf("/api/coinflip/%d", gamesresp[0].ID), jwt2, "Game Completed or Same PlayerId")
	callErrEndPoint(t, nil, "PUT", fmt.Sprintf("/api/coinflip/%d", gamesresp[1].ID), jwt1, "Game Completed or Same PlayerId")
	callErrEndPoint(t, nil, "DELETE", fmt.Sprintf("/api/coinflip/%d", gamesresp[0].ID), jwt1, "sql: no rows in result set")
	callErrEndPoint(t, nil, "DELETE", fmt.Sprintf("/api/coinflip/%d", gamesresp[1].ID), jwt2, "sql: no rows in result set")

	_, body = callEndPoint(t, ts, "GET", "/api/coinflip", nil, "")
	assert.Equal(t, "[]", strings.TrimSpace(body), body)

	// Test Delete
	callOkEndPoint(t, game1, "POST", "/api/coinflip", jwt1, "Game Created")
	callOkEndPoint(t, game2, "POST", "/api/coinflip", jwt2, "Game Created")
	_, body = callEndPoint(t, ts, "GET", "/api/coinflip", nil, "")
	gamesresp = make([]coinflip.GameOpen, 2)
	validateGamesList(t, body, gamesresp, 2)

	callErrEndPoint(t, nil, "DELETE", fmt.Sprintf("/api/coinflip/%d", gamesresp[0].ID), jwt2, "sql: no rows in result set")
	callErrEndPoint(t, nil, "DELETE", fmt.Sprintf("/api/coinflip/%d", gamesresp[1].ID), jwt1, "sql: no rows in result set")

	b1i = getBalance(testUserID1)
	callOkEndPoint(t, nil, "DELETE", fmt.Sprintf("/api/coinflip/%d", gamesresp[0].ID), jwt1, "Game Deleted")
	b1 = getBalance(testUserID1)
	assert.Equal(t, b1i+float64(game1.Amount), b1)

	b2i := getBalance(testUserID2)
	callOkEndPoint(t, nil, "DELETE", fmt.Sprintf("/api/coinflip/%d", gamesresp[1].ID), jwt2, "Game Deleted")
	b2 = getBalance(testUserID2)
	assert.Equal(t, b2i+float64(game2.Amount), b2)

	_, body = callEndPoint(t, ts, "GET", "/api/coinflip", nil, "")
	assert.Equal(t, "[]", strings.TrimSpace(body), body)

	// test user history
	_, body = callEndPoint(t, ts, "GET", "/api/coinflip-history", nil, jwt1)
	gamesrespC := make([]coinflip.GameCompleted, 2)
	validateGamesListC(t, body, gamesrespC, 2)
	assert.Equal(t, game1.Amount, gamesrespC[1].Amount, body)
	assert.Equal(t, testUserID1, gamesrespC[1].Player1Id, body)
	assert.Equal(t, float64(19.5), gamesrespC[1].Payout, body)
	assert.Equal(t, float64(0.5), gamesrespC[1].Commision, body)

	_, body = callEndPoint(t, ts, "GET", "/api/coinflip-history", nil, jwt2)
	gamesrespC = make([]coinflip.GameCompleted, 2)
	validateGamesListC(t, body, gamesrespC, 2)
	assert.Equal(t, game2.Amount, gamesrespC[0].Amount, body)
	assert.Equal(t, testUserID2, gamesrespC[0].Player1Id, body)
	assert.Equal(t, float64(39), gamesrespC[0].Payout, body)
	assert.Equal(t, float64(1), gamesrespC[0].Commision, body)

	// test top players
	_, body = callEndPoint(t, ts, "GET", "/api/coinflip-top-players", nil, "")
	playersresp := make([]coinflip.TopPlayer, 2)
	validateTopPlayers(t, body, playersresp, 2)
	assert.Equal(t, uint(20), playersresp[0].Amount, body)
	assert.Equal(t, uint(10), playersresp[1].Amount, body)
}
