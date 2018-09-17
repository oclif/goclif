package main

import (
	"os"
)

func main() {
	if os.Args[1] == "__goclifd" {
		daemon()
	} else {
		Run(os.Args[1:])
	}
}

// Run runs a command
func Run(argv []string) {
	run(argv)
}
