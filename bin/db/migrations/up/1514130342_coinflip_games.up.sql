CREATE TABLE coinflip_games
(
  id         BIGSERIAL                                                                NOT NULL
    CONSTRAINT coinflip_games_pkey
    PRIMARY KEY,
  player1_id BIGINT                                                                   NOT NULL
    CONSTRAINT coinflip_games_player1_id_foreign
    REFERENCES users,
  player2_id BIGINT
    CONSTRAINT coinflip_games_player2_id_foreign
    REFERENCES users,
  amount     INTEGER                                                                  NOT NULL,
  commision  NUMERIC(14, 4),
  payout     NUMERIC(14, 4),
  pick       INTEGER                                                                  NOT NULL,
  result     INTEGER                                                                  NOT NULL,
  hash       VARCHAR(255)                                                             NOT NULL,
  salt       VARCHAR(255)                                                             NOT NULL,
  started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP                       NOT NULL
);
CREATE INDEX coinflip_games_player1_id_index
  ON coinflip_games (player1_id);
CREATE INDEX coinflip_games_player2_id_index
  ON coinflip_games (player2_id);  