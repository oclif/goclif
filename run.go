package main

import (
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"time"
)

func connect(retry bool) net.Conn {
	debugf("connecting")
	c, err := net.DialTimeout("unix", socketOrchestrator, time.Second*5)
	if err != nil && retry {
		forkDaemon()
		return connect(false)
	}
	must(err)
	send(c, MessageInit{"init", rand.Intn(100000)})
	return c
}

func send(c io.Writer, msg interface{}) {
	output, err := json.Marshal(msg)
	debugf("sent: %s\n", string(output))
	must(err)
	c.Write(output)
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

	send(MessageCommand{
		Type: "command",
		Argv: argv,
	})

	read()
}
