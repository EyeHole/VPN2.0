package client

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	commands "VPN2.0/cmd"
)

const (
	serverAddr = "localhost:8080"
)

func copyTo(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}

func requestCreateNetwork() error {
	conn, _ := net.Dial("tcp", serverAddr)
	go copyTo(os.Stdout, conn)
	_, err := conn.Write([]byte(commands.CreateCmd))
	if err != nil {
		return err
	}
	fmt.Println("sent")
	return nil
}

func processCmd(cmd string) error {
	switch cmd {
	case commands.CreateCmd:
		return requestCreateNetwork()
	default:
		return errors.New("undefined cmd")
	}
}

func RunClient() error {
	fmt.Println("Enter cmd:")
	var cmd string
	for {
		_, err := fmt.Scanf("%s", &cmd)
		if err != nil {
			return err
		}
		if err := processCmd(cmd); err != nil {
			return err
		}
	}
}
