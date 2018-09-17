package main

import (
	"log"
	"os"
)

var (
	debugging = os.Getenv("DEBUG") != ""
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func debugf(f string, args ...interface{}) {
	if debugging {
		log.Printf(f, args...)
	}
}
