package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

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
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Fatal("read error: ", err)
		}
		debugf("client got: %#v\n", string(buf[0:n]))
	}
}

func main() {
	c, err := net.Dial("unix", "/tmp/foo.sock")
	if err != nil {
		log.Fatal("dial error: ", err)
	}
	defer c.Close()

	send := func(msg string) {
		debugf("client sent: %#v\n", msg)
		c.Write([]byte(msg))
	}

	send("version")

	read(c)
}
