package config

import "os"

var DEBUG = false

func init() {
	debug := os.Getenv("DEBUG")
	if debug == "1" {
		DEBUG = true
	}
}
