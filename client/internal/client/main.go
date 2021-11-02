package client

import (
	commands "VPN2.0/lib/cmd"
	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/tap"
	"bufio"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net"
)

const (
	serverAddr = "localhost:8080"
)

func copyTo(ctx context.Context, dst io.Writer, src io.Reader) {
	logger := ctxmeta.GetLogger(ctx)

	_, err := io.Copy(dst, src);
	if err != nil {
		logger.Error("error in copyTo", zap.Error(err))
	}
}

func processResp(ctx context.Context, conn net.Conn, cmdName string, errCh chan error){
	logger := ctxmeta.GetLogger(ctx)

	clientReader := bufio.NewReader(conn)
	resp, err := clientReader.ReadString('\n')
	if err != nil {
		logger.Error("failed to read from conn", zap.Error(err))
		errCh <- err
	}

	logger.Info("Got resp", zap.String("resp", resp))

	switch cmdName {
	case commands.ConnectCmd:
		respStrings := commands.GetWords(resp)
		if len(respStrings) < 2 {
			logger.Error("empty resp from server")
			errCh <- errors.New("empty resp")
		}
		if respStrings[0] != commands.SuccessResponse {
			logger.Error("got error in server resp")
			errCh <- errors.New("error in resp")
		}

		tapName := tap.GetTapName("client", 1, rand.Int())
		_, err := tap.ConnectToTap(ctx, tapName)
		if err != nil {
			errCh <- err
		}

		err = tap.SetTapUp(ctx, respStrings[1], "client_tap1")
		if err != nil {
			errCh <- err
		}
	}
}

func makeRequest(ctx context.Context, msg string, cmdName string) (err error) {
	logger := ctxmeta.GetLogger(ctx)

	conn, _ := net.Dial("tcp", serverAddr)

	errCh := make(chan error, 1)
	go processResp(ctx, conn, cmdName, errCh)
	close(errCh)
	if err := <-errCh; err != nil {
		return err
	}

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
	return makeRequest(ctx, msg, commands.CreateCmd)
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
	return makeRequest(ctx, msg, commands.ConnectCmd)
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
