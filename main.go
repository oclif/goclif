package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
)

type commandMessage struct {
	Type string   `json:"type"`
	Argv []string `json:"argv"`
}

var (
	debugging = os.Getenv("DEBUG") != ""
)

func debugf(msg string, args ...interface{}) {
	if debugging {
		fmt.Printf(msg, args...)
	}
}

func read(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err == io.EOF {
			return
		}
		must(err)
		debugf("client got: %#v\n", string(buf[0:n]))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	c, err := net.Dial("unix", "/tmp/foo.sock")
	must(err)
	defer c.Close()

	send := func(msg interface{}) {
		output, err := json.Marshal(msg)
		debugf("client sent: %s\n", string(output))
		must(err)
		c.Write(output)
	}

	send(commandMessage{
		Type: "command",
		Argv: []string{"version"},
	})

	read(c)
}
