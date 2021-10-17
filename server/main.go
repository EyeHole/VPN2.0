package server

import (
	"fmt"
	"io"
	"net"

	commands "github.com/VPN2.0/cmd"
)

const (
	readLimit = 2048
	serverAddr = "localhost:8080"
)

func runServer() error {
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		var errCh chan error
		go handleClient(conn, errCh)
	}
}

func createNetwork(conn net.Conn) error {
	_, err := conn.Write([]byte("Got your request"))
	return err
}

func processCmd(cmd string, conn net.Conn) error {
	switch cmd {
	case commands.CreateCmd:
		return createNetwork(conn)
	}

	return nil
}

func handleClient(conn net.Conn, errCh chan error) {
	defer func() {
		err := conn.Close()
		if err != nil {
			errCh <- err
		}
	}()

	var cmd string
	buf := make([]byte, 32)
	for len(cmd) < readLimit {
		_, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				errCh <- err
			}
			break
		}
		cmd += string(buf)
	}

	err := processCmd(cmd, conn)
	if err != nil {
		errCh <- err
	}
}

func serverCreate() {
}

func main() {
	serverCreate()

	err := runServer()
	if err != nil {
		fmt.Println(err)
	}
}
