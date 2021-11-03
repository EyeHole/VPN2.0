package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"

	commands "VPN2.0/lib/cmd"
	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/tap"
)

const (
	serverAddr = "localhost:8080"
)

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

		rand.Seed(time.Now().UnixNano())
		tapName := tap.GetTapName("client", 1, 10 + rand.Intn(191))
		tapIf, err := tap.ConnectToTap(ctx, tapName)
		if err != nil {
			errCh <- err
		}
		logger.Debug("connected to tap", zap.String("tap_name", tapName))

		err = tap.SetTapUp(ctx, respStrings[1], tapName)
		if err != nil {
			errCh <- err
		}
		logger.Debug("set tap up", zap.String("tap_name", tapName))


		go tap.HandleTapEvent(ctx, tapIf, conn, errCh)
		go tap.HandleConnEvent(ctx, tapIf, conn, errCh)
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
