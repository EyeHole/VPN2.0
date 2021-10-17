package server

import (
	commands "VPN2.0/cmd"
	"fmt"
	"io"
	"net"
	"strconv"
)

const (
	bufferSize = 32
	readLimit  = 2048
	serverAddr = "localhost:8080"
)

func RunServer() error {
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		fmt.Println("hello stranger")
		if err != nil {
			fmt.Println(":(")
			continue
		}
		errCh := make(chan error)
		go handleClient(conn, errCh)
		select {
		case err := <-errCh:
			return err
		}

	}
}

func createNetwork(conn net.Conn) error {
	fmt.Println("got something")
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

	for {
		buf := make([]byte, 32)
		readLen, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("wowowowowo")
				errCh <- err
			} else {
				fmt.Println("got EOF")
			}
			break
		}
		if readLen == 0 {
			fmt.Println("wow")
		}
		fmt.Println("readLen: " + strconv.Itoa(readLen))
		cmd += string(buf)
	}

	//_, err := fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	//if err != nil {
	//	errCh <- err
	//}
	//buf, err := ioutil.ReadAll(conn)
	//if err != nil {
	//	errCh <- err
	//}
	//fmt.Println("hey")
	//cmd := string(buf)
	//
	//fmt.Println("got cmd:" + cmd)
	err := processCmd(cmd, conn)
	if err != nil {
		errCh <- err
	}
}

func CreateServer() {
}
