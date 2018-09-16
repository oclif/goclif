package main

import (
	"fmt"
	"os"
)

var (
	debugging = os.Getenv("DEBUG") != ""
)

func debugf(msg string, args ...interface{}) {
	if debugging {
		fmt.Printf("client "+msg, args...)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
