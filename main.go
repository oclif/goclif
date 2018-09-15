package main

import (
	"io"
	"log"
	"net"
)

func read(r io.Reader) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if err != nil {
			log.Fatal("read error: ", err)
		}
		println("client got:", string(buf[0:n]))
	}
}

func main() {
	c, err := net.Dial("unix", "/tmp/foo.sock")
	if err != nil {
		log.Fatal("dial error: ", err)
	}
	defer c.Close()

	send := func(msg string) {
		println("client sent:", msg)
		c.Write([]byte(msg))
	}

	send("hi there bud")

	read(c)
}
