package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	User "os/user"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/erikdubbelboer/gspt"
)

func getUser() string {
	user, err := User.Current()
	must(err)
	return user.Username
}

var (
	user               = getUser()
	socketBase         = path.Join(os.TempDir(), user, "goclifd.sock")
	socketOrchestrator = path.Join(socketBase, "orchestrator")
)

func socketRun(id int) string {
	return path.Join(socketBase, "run-"+strconv.Itoa(id))
}

func daemon() {
	debugf("started daemon\n")
	gspt.SetProcTitle("goclifd")
	must(os.RemoveAll(socketBase))
	os.MkdirAll(socketBase, 0700)
	s, err := net.Listen("unix", socketOrchestrator)
	must(err)
	connections := make(chan net.Conn, 1)
	go func() {
		for {
			c, err := s.Accept()
			must(err)
			connections <- c
		}
	}()
	workers := []bool{}
	startWorker := func() {
		id := len(workers)
		debugf("starting worker %d\n", id)
		socketBase := socketRun(id)
		os.MkdirAll(socketBase, 0700)
		cmd := exec.Command("node", "-", "--", socketBase)
		cmd.Stdin = strings.NewReader(MustAssetString("server.js"))
		cmd.Stderr = os.Stderr
		must(cmd.Run())
		workers = append(workers, true)
	}
	go startWorker()
	handle := func(c net.Conn) {
		send(c, MessageInit{"init", 0})
	}
	fmt.Println(socketOrchestrator)
	for {
		select {
		case c := <-connections:
			debugf("daemon received connection")
			go handle(c)
		case <-time.After(time.Second * 5):
			debugf("closing")
			must(os.RemoveAll(socketBase))
			return
		}
	}
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
