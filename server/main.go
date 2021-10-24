package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"

	"github.com/google/uuid"

	commands "VPN2.0/cmd"
)

const (
	cmdLenLimit = 32
	serverAddr  = "localhost:8080"
)

func createBridge(networkID string) error {
	bridgeName := fmt.Sprintf("b-%s", networkID)

	_, err := exec.Command("ip", "link", "add", "name", bridgeName, "type", "bridge").Output()
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	fmt.Printf("bridge %s created!\n", bridgeName)

	_, err = exec.Command("ip", "link", "set", bridgeName, "up").Output()
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("bridge %s is up!\n", bridgeName)

	return nil
}

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

		errCh := make(chan error, 1)
		go handleClient(conn, errCh)

		close(errCh)
		if err := <-errCh; err != nil {
			return err
		}
	}
}

func createNetwork(conn net.Conn) error {
	fmt.Println("got something")

	netID := uuid.New().String()[:5]
	err := createBridge(netID)

	resp := "network created"
	if err != nil {
		resp = "failed to process request"
	}
	_, err = conn.Write([]byte(resp))
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

	clientReader := bufio.NewReader(conn)
	for {
		cmd, err := clientReader.ReadString('\n')
		switch err {
		case nil:
			cmd := strings.TrimSpace(cmd)
			fmt.Println(cmd)

			err = processCmd(cmd, conn)
			if err != nil {
				errCh <- err
			}
		case io.EOF:
			fmt.Println("client closed the connection by terminating the process")
			return
		default:
			fmt.Printf("error: %v\n", err)
			errCh <- err
			return
		}
	}
}

func CreateServer() {
}
