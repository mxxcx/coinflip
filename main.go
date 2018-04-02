package main

import (
	"log"
	"net/http"

	"github.com/mtdx/ns-ga/coinflip"

	"github.com/mtdx/ns-ga/db"
	"github.com/mtdx/ns-ga/rest"
)

func main() {
	dbconn := db.Open()
	db.RunMigrations(dbconn)
	defer dbconn.Close()

	go coinflip.BroadcastWs()

	r := rest.Router(dbconn)
	err := http.ListenAndServe(":5000", r)
	log.Fatal(err.Error())
}
