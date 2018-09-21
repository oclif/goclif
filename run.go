package main

import (
	"encoding/json"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

type connection struct {
	ctl    net.Conn
	stdin  net.Conn
	stdout net.Conn
	stderr net.Conn
}

func send(c net.Conn, msg interface{}) {
	output, err := json.Marshal(msg)
	debugf("sent: %s\n", string(output))
	must(err)
	c.Write(output)
}

func getMessage(uuid string, c net.Conn) *Message {
	var msg Message
	decoder := json.NewDecoder(c)
	must(decoder.Decode(&msg))
	if msg.ID != uuid {
		debugf("sending back msg %#v\n", msg)
		// send it back
		send(c, msg)
		return getMessage(uuid, c)
	}
	debugf("got: %#v\n", msg)
	return &msg
}

func connect(argv []string, retry bool) *connection {
	debugf("connecting")
	orchestrator, err := net.DialTimeout("unix", socketOrchestrator, time.Second*5)
	if err != nil && retry {
		forkDaemon()
		return connect(argv, false)
	}
	must(err)
	u := uuid.Must(uuid.NewV4()).String()
	send(orchestrator, Message{u, nil, "command", argv})
	msg := getMessage(u, orchestrator)
	stdin, err := net.DialTimeout("unix", socketRun(*msg.WorkerID, "stdin"), time.Second*5)
	must(err)
	stdout, err := net.DialTimeout("unix", socketRun(*msg.WorkerID, "stdout"), time.Second*5)
	must(err)
	stderr, err := net.DialTimeout("unix", socketRun(*msg.WorkerID, "stderr"), time.Second*5)
	must(err)
	return &connection{orchestrator, stdin, stdout, stderr}
}

func run(argv []string) {
	c := connect(argv, true)
	debugf("connected")
	defer c.ctl.Close()
	defer c.stdin.Close()
	defer c.stdout.Close()
	defer c.stderr.Close()

	read := func(stream string, c net.Conn) {
		buf := make([]byte, 1024)
		for {
			n, err := c.Read(buf[:])
			if err != nil {
				if err == io.EOF {
					return
				}
				if strings.HasSuffix(err.Error(), "use of closed network connection") {
					return
				}
				panic(err)
			}
			debugf("got: %#v\n", string(buf[0:n]))
			switch stream {
			case "stdout":
				os.Stdout.Write(buf)
			case "stderr":
				os.Stderr.Write(buf)
			}
		}
	}

	go read("stdout", c.stdout)
	go read("stderr", c.stderr)
	decoder := json.NewDecoder(c.ctl)
	var exit struct {
		Code int `json:"code"`
	}
	must(decoder.Decode(&exit))
	c.ctl.Close()
	c.stdin.Close()
	c.stdout.Close()
	c.stderr.Close()
	os.Exit(exit.Code)
}
