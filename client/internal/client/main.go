package client

import (
	"VPN2.0/lib/ctxmeta"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net"
	"os"

	commands "VPN2.0/cmd"
)

const (
	serverAddr = "localhost:8080"
)

func copyTo(ctx context.Context, dst io.Writer, src io.Reader) {
	logger := ctxmeta.GetLogger(ctx)

	if _, err := io.Copy(dst, src); err != nil {
		logger.Error("error in copyTo", zap.Error(err))
	}
}

func makeRequest(ctx context.Context, msg string) (err error) {
	logger := ctxmeta.GetLogger(ctx)

	conn, _ := net.Dial("tcp", serverAddr)
	go copyTo(ctx, os.Stdout, conn)

	_, err = conn.Write([]byte(msg + "\n"))
	if err != nil {
		logger.Error("failed to write to conn", zap.Error(err))
		return err
	}

	logger.Debug("sent cmd")

	return nil
}

func processCreateRequest(ctx context.Context) error {
	logger := ctxmeta.GetLogger(ctx)

	var name, password string
	fmt.Println("Enter net name:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		logger.Error("got error while scanning create net name", zap.Error(err))
		return err
	}

	fmt.Println("Enter net password:")
	_, err = fmt.Scanf("%s", &password)
	if err != nil {
		logger.Error("got error while scanning create net password", zap.Error(err))
		return err
	}

	msg := fmt.Sprintf("%s %s %s", commands.CreateCmd, name, password)
	return makeRequest(ctx, msg)
}

func processConnectRequest(ctx context.Context) error {
	logger := ctxmeta.GetLogger(ctx)

	var name, password string
	fmt.Println("Enter net name:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		logger.Error("got error while scanning create net name", zap.Error(err))
		return err
	}

	fmt.Println("Enter net password:")
	_, err = fmt.Scanf("%s", &password)
	if err != nil {
		logger.Error("got error while scanning create net password", zap.Error(err))
		return err
	}

	msg := fmt.Sprintf("%s %s %s", commands.ConnectCmd, name, password)
	return makeRequest(ctx, msg)
}

func processCmd(ctx context.Context, cmd string) error {
	logger := ctxmeta.GetLogger(ctx)

	switch cmd {
	case commands.CreateCmd:
		return processCreateRequest(ctx)
	case commands.ConnectCmd:
		return processConnectRequest(ctx)
	default:
		logger.Error("undefined cmd")
		return errors.New("undefined cmd")
	}
}

func RunClient(ctx context.Context) error {
	logger := ctxmeta.GetLogger(ctx)
	fmt.Println("Enter cmd:")

	var cmd string
	for {
		_, err := fmt.Scanf("%s", &cmd)
		if err != nil {
			logger.Error("got error while scanning cmd", zap.Error(err))
			return err
		}

		if err := processCmd(ctx, cmd); err != nil {
			return err
		}
	}
}
