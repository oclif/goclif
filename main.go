package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
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
		fmt.Printf("client "+msg, args...)
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
		debugf("got: %#v\n", string(buf[0:n]))
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func startDaemon() string {
	debugf("starting daemon\n")
	cmd := exec.Command("node")
	cmd.Stdin = strings.NewReader(MustAssetString("server.js"))
	stdoutRaw, err := cmd.StdoutPipe()
	must(err)
	stdout := bufio.NewReader(stdoutRaw)
	cmd.Stderr = os.Stderr
	must(cmd.Start())
	debugf("started daemon\n")
	go func() {
		must(cmd.Wait())
	}()
	socket, err := stdout.ReadString('\n')
	socket = strings.TrimSpace(socket)
	debugf("socket: %s\n", socket)
	must(err)
	return socket
}

// Run runs the CLI
func Run(argv []string) {
	socket := startDaemon()
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

func main() {
	Run(os.Args[1:])
}
