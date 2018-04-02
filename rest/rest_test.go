package rest

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mtdx/ns-ga/coinflip"

	"github.com/go-chi/jwtauth"
	"github.com/mtdx/ns-ga/common"
	"github.com/mtdx/ns-ga/config"
	"github.com/mtdx/ns-ga/db"
	"github.com/mtdx/ns-ga/validator"
	"github.com/stretchr/testify/assert"
)

const testID = "11111111111111111"
const testUserID1 = "76561198181937286"
const testUserID2 = "76561198449205177"

var ts *httptest.Server
var body, jwt1, jwt2 string
var jsonreq []byte
var err error
var dbconn *sql.DB

func TestMain(m *testing.M) {
	dbconn = db.Open()
	defer dbconn.Close()

	r := Router(dbconn)
	ts = httptest.NewServer(r)
	defer ts.Close()

	cleanTestUserData()
	setupTestUserData()

	code := m.Run() // run tests & exit
	os.Exit(code)
}

func TestGameCreateAuth(t *testing.T) {
	t.Parallel()
	assertAuth(t, ts, "POST", "/api/coinflip")
	assertAuth(t, ts, "GET", "/api/coinflip-history")
}

func TestCoinflipGame(t *testing.T) {
	t.Parallel()
	coinflipGameTest(t)
}

func setupTestUserData() {
	tokenAuth := jwtauth.New("HS256", []byte(config.JwtKey()), nil)
	exp := time.Now().Add(time.Hour * time.Duration(24)).Unix()
	claims1 := jwtauth.Claims{"steam_id": testUserID1, "exp": exp, "referral_code": nil, "iat": time.Now().Unix()}
	claims2 := jwtauth.Claims{"steam_id": testUserID2, "exp": exp, "referral_code": nil, "iat": time.Now().Unix()}
	_, token1, _ := tokenAuth.Encode(claims1)
	_, token2, _ := tokenAuth.Encode(claims2)

	_, err := dbconn.Exec(`INSERT INTO users (steam_id, trade_link_url, balance, access_code, balance_withheld, terms_confirmed, last_balance_change_reason, daily_bonus_available_at, _total_wagered, _total_deposited, _total_withdrawn, is_banned, is_moderator, is_admin, referral_code, created_at, referral_code_used, referrer_steam_id, referee_bonus_claimed_at, last_login_at, last_login_ip, withdrawals_disabled, notify_referral_code_changed, daily_bonus_wager_requirement) 
	VALUES ($1, '', 1000.0000, null, 100, true, null, '2017-12-18 19:37:14.786152', 80, 0, 0, false, false, false, null, '2017-12-18 19:37:14.786152', null, null, null, '2017-12-26 14:46:43.228142', '127.0.0.1', false, false, 0.00) ON CONFLICT (steam_id) DO UPDATE SET balance = 1000; `, testUserID1)
	_, err = dbconn.Exec(`INSERT INTO users (steam_id, trade_link_url, balance, access_code, balance_withheld, terms_confirmed, last_balance_change_reason, daily_bonus_available_at, _total_wagered, _total_deposited, _total_withdrawn, is_banned, is_moderator, is_admin, referral_code, created_at, referral_code_used, referrer_steam_id, referee_bonus_claimed_at, last_login_at, last_login_ip, withdrawals_disabled, notify_referral_code_changed, daily_bonus_wager_requirement) 
	VALUES ($1, null, 2000.0000, null, 200, true, null, '2017-12-18 19:49:16.752971', 0, 0, 0, false, false, false, null, '2017-11-18 19:49:16.752000', null, null, null, '2017-12-18 19:49:17.278428', '127.0.0.1', false, false, 0.00) ON CONFLICT (steam_id) DO UPDATE SET balance = 2000`, testUserID2)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add test data: %v\n", err)
		os.Exit(1)
	}

	jwt1, jwt2 = token1, token2
}

func callOkEndPoint(t *testing.T, data interface{}, method string, endpoint string, jwt string, result string) {
	var successResp common.SuccessResponse
	jsonreq, _ := json.Marshal(data)
	_, body := callEndPoint(t, ts, method, endpoint, bytes.NewReader(jsonreq), jwt)
	if err := json.Unmarshal([]byte(body), &successResp); err != nil {
		t.Fatalf("Failed to Unmarshal, got: %s, error: %s", body, err.Error())
	}
	assert.Equal(t, result, successResp.StatusText, body)
}

func validateGamesList(t *testing.T, body string, resp []coinflip.GameOpen, minlen int) {
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("Failed to Unmarshal, got: %s, error: %s", body, err.Error())
	}
	for _, res := range resp {
		if err := validator.Validate(res); err != nil {
			t.Fatalf("Error: %s", err.Error())
		}
	}
	assert.Equal(t, minlen, len(resp), body)
}

func validateGamesListC(t *testing.T, body string, resp []coinflip.GameCompleted, minlen int) {
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("Failed to Unmarshal, got: %s, error: %s", body, err.Error())
	}
	for _, res := range resp {
		if err := validator.Validate(res); err != nil {
			t.Fatalf("Error: %s", err.Error())
		}
	}
	assert.Equal(t, minlen, len(resp), body)
}

func validateTopPlayers(t *testing.T, body string, resp []coinflip.TopPlayer, minlen int) {
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("Failed to Unmarshal, got: %s, error: %s", body, err.Error())
	}
	for _, res := range resp {
		if err := validator.Validate(res); err != nil {
			t.Fatalf("Error: %s", err.Error())
		}
	}
	assert.Equal(t, minlen, len(resp), body)
}

func callErrEndPoint(t *testing.T, data interface{}, method string, endpoint string, jwt string, result string) {
	var errResp common.ErrResponse
	jsonreq, _ := json.Marshal(data)
	_, body := callEndPoint(t, ts, method, endpoint, bytes.NewReader(jsonreq), jwt)
	if err := json.Unmarshal([]byte(body), &errResp); err != nil {
		t.Fatalf("Failed to Unmarshal, got: %s, error: %s", body, err.Error())
	}
	assert.Equal(t, result, errResp.StatusText, body)
}

func cleanTestUserData() {
	_, err = dbconn.Exec(`DELETE FROM coinflip_games WHERE player1_id = $1 OR player1_id = $2`, testUserID1, testUserID2)
	_, err = dbconn.Exec(`DELETE FROM users WHERE steam_id = $1 OR steam_id = $2`, testUserID1, testUserID2)
	_, err = dbconn.Exec(`DELETE FROM user_balance_history WHERE user_steam_id = $1 OR user_steam_id = $2`, testUserID1, testUserID2)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func getBalance(userID string) float64 {
	var balance float64
	err := dbconn.QueryRow("SELECT balance FROM users WHERE steam_id = $1", userID).Scan(&balance)
	if err != nil || err == sql.ErrNoRows {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	return balance
}

func assertAuth(t *testing.T, ts *httptest.Server, method string, route string) {
	_, body := callEndPoint(t, ts, method, route, nil, "jwt-test-xyz")
	assert.Equal(t, `Unauthorized`, strings.TrimSpace(body), body)
}

func callEndPoint(t *testing.T, ts *httptest.Server, method, path string, body io.Reader, jwt string) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	if jwt != "" {
		req.Header.Add("Authorization", "BEARER "+jwt)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return nil, ""
	}
	defer resp.Body.Close()

	return resp, string(respBody)
}
