package main

import (
	"encoding/json"
	"io"
	"net"
	"time"
)

func connect(retry bool) net.Conn {
	debugf("connecting")
	c, err := net.DialTimeout("unix", socket, time.Second*5)
	if err != nil {
		if err.Error() == "dial unix "+socket+": connect: no such file or directory" {
			forkDaemon()
			return connect(false)
		}
		must(err)
	}
	return c
}

func run(argv []string) {
	c := connect(true)
	debugf("connected")
	defer c.Close()

	read := func() {
		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf[:])
			if err == io.EOF {
				return
			}
			must(err)
			debugf("got: %#v\n", string(buf[0:n]))
		}
	}

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

	read()
}
