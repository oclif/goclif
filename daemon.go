package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/erikdubbelboer/gspt"
)

var (
	socket = path.Join(os.TempDir(), "goclifd.sock")
)

func daemon() {
	must(os.RemoveAll(socket))
	gspt.SetProcTitle("goclifd")
	debugf("started daemon\n")
	s, err := net.Listen("unix", socket)
	must(err)
	fmt.Println(socket)
	connections := make(chan net.Conn, 1)
	go func() {
		for {
			c, err := s.Accept()
			must(err)
			connections <- c
		}
	}()
	for {
		select {
		case <-connections:
			debugf("daemon received connection")
		case <-time.After(time.Second * 5):
			debugf("closing")
			must(os.RemoveAll(socket))
			return
		}
	}
	// cmd := exec.Command("node")
	// cmd.Stdin = strings.NewReader(MustAssetString("server.js"))
	// stdoutRaw, err := cmd.StdoutPipe()
	// must(err)
	// stdout := bufio.NewReader(stdoutRaw)
	// cmd.Stderr = os.Stderr
	// must(cmd.Start())
	// socket, err := stdout.ReadString('\n')
	// socket = strings.TrimSpace(socket)
	// debugf("socket: %s\n", socket)
	// must(err)
	// return socket
}

func forkDaemon() string {
	debugf("starting daemon\n")
	readFirstLine := func(cmd *exec.Cmd) string {
		stdoutRaw, err := cmd.StdoutPipe()
		must(err)
		stdout := bufio.NewReader(stdoutRaw)
		must(cmd.Start())
		s, err := stdout.ReadString('\n')
		must(err)
		return strings.TrimSpace(s)
	}
	cmd := exec.Command("./dist/goclif", "__goclifd")
	cmd.Stderr = os.Stderr
	return readFirstLine(cmd)
}
