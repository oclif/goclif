package main

import (
	"bufio"
	"encoding/json"
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

func socketRun(id int, stream string) string {
	return path.Join(socketBase, strconv.Itoa(id), stream)
}

type worker struct {
	ID      int
	Working bool
	Cmd     *exec.Cmd
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
	workers := []*worker{}
	startWorker := func() *worker {
		id := len(workers)
		worker := worker{id, true, nil}
		workers = append(workers, &worker)
		debugf("starting worker %d\n", id)
		socketBase := socketRun(id, "")
		os.MkdirAll(socketBase, 0700)
		worker.Cmd = exec.Command("node", "-", "--", socketBase)
		worker.Cmd.Stdin = strings.NewReader(MustAssetString("server.js"))
		worker.Cmd.Stderr = os.Stderr
		readFirstLine(worker.Cmd)
		worker.Working = false
		return &worker
	}
	getWorker := func() *worker {
		for _, worker := range workers {
			debugf("%#v\n", worker)
			if !worker.Working {
				return worker
			}
		}
		return startWorker()
	}
	handle := func(c net.Conn) {
		decoder := json.NewDecoder(c)
		var msg Message
		must(decoder.Decode(&msg))
		debugf("got: %#v\n", msg)
		worker := getWorker()
		msg.WorkerID = &worker.ID
		worker.Working = true
		ctl, err := net.DialTimeout("unix", socketRun(worker.ID, "ctl"), time.Second*5)
		must(err)
		defer ctl.Close()
		send(ctl, msg)
		send(c, msg)
		decoder = json.NewDecoder(ctl)
		var exit struct {
			Code int `json:"code"`
		}
		must(decoder.Decode(&exit))
		send(c, exit)
		worker.Working = false
		debugf("worker done")
	}
	fmt.Println(socketOrchestrator)
	for {
		select {
		case c := <-connections:
			debugf("daemon received connection")
			go handle(c)
		case <-time.After(time.Second * 10):
			debugf("closing")
			must(os.RemoveAll(socketBase))
			for _, worker := range workers {
				worker.Cmd.Process.Kill()
			}
			return
		}
	}
}

func readFirstLine(cmd *exec.Cmd) string {
	stdoutRaw, err := cmd.StdoutPipe()
	must(err)
	stdout := bufio.NewReader(stdoutRaw)
	must(cmd.Start())
	s, err := stdout.ReadString('\n')
	must(err)
	return strings.TrimSpace(s)
}

func forkDaemon() string {
	debugf("starting daemon\n")
	cmd := exec.Command("./dist/goclif", "__goclifd")
	cmd.Stderr = os.Stderr
	return readFirstLine(cmd)
}
