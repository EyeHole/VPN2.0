package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	commands "VPN2.0/cmd"
)

const (
	readLimit = 2048
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
		case err := <- errCh:
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

	//var cmd string
	//buf := make([]byte, 32)
	//for len(cmd) < readLimit {
	//	readLen, err := conn.Read(buf)
	//	if err != nil {
	//		if err != io.EOF {
	//			errCh <- err
	//		} else {
	//			fmt.Println("got EOF")
	//		}
	//		break
	//	}
	//	fmt.Println("readLen: " + strconv.Itoa(readLen))
	//	cmd += string(buf)
	//}

	stringmy := ""
	readLen, err := fmt.Fprintf(conn, "%s", stringmy)
	fmt.Println("readLen: " + strconv.Itoa(readLen) + ", str: " + stringmy)
	buf, err := ioutil.ReadAll(conn)
	if err != nil {
		errCh <- err
	}

	cmd := string(buf)

	fmt.Println("got cmd:" + cmd)
	err = processCmd(cmd, conn)
	if err != nil {
		errCh <- err
	}
}

func CreateServer() {
}
