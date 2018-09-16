package main

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

func daemon() string {
	debugf("starting daemon\n")
	cmd := exec.Command("node")
	cmd.Stdin = strings.NewReader(MustAssetString("server.js"))
	stdoutRaw, err := cmd.StdoutPipe()
	must(err)
	stdout := bufio.NewReader(stdoutRaw)
	cmd.Stderr = os.Stderr
	must(cmd.Start())
	debugf("started daemon\n")
	socket, err := stdout.ReadString('\n')
	socket = strings.TrimSpace(socket)
	debugf("socket: %s\n", socket)
	must(err)
	return socket
}
