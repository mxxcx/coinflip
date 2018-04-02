package config

import "os"

// Env config from env or dev defaults
func Env() map[string]string {
	m := map[string]string{
		"host":     os.Getenv("NS_DB_HOST"),
		"user":     os.Getenv("NS_DB_USER"),
		"password": os.Getenv("NS_DB_PASSWORD"),
		"database": os.Getenv("NS_DB_DATABASE"),
	}

	if m["host"] == "" {
		m["host"] = "localhost"
	}
	if m["user"] == "" {
		m["user"] = "user"
	}
	if m["password"] == "" {
		m["password"] = "password"
	}
	if m["database"] == "" {
		m["database"] = "dbname"
	}

	return m
}
