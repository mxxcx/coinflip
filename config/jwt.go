package config

import "os"

// JwtKey private key from env or insecure default
func JwtKey() string {
	key := os.Getenv("NS_ADMIN_SIGNING_KEY")
	if key == "" {
		key = "InsecurePrivateKey" // default insecure private key
	}
	return key
}
