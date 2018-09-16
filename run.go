package main

import (
	"encoding/json"
	"io"
	"net"
)

func read(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err == io.EOF {
			return
		}
		must(err)
		debugf("got: %#v\n", string(buf[0:n]))
	}
}

// Run runs the CLI
func Run(argv []string) {
	socket := daemon()
	c, err := net.Dial("unix", socket)
	must(err)
	defer c.Close()

	send := func(msg interface{}) {
		output, err := json.Marshal(msg)
		debugf("sent: %s\n", string(output))
		must(err)
		c.Write(output)
	}

	send(commandMessage{
		Type: "command",
		Argv: argv,
	})

	read(c)
}
