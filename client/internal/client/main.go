package client

import (
	"VPN2.0/lib/localnet"
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
	"VPN2.0/lib/tun"
)

func processResp(ctx context.Context, conn net.Conn, cmdName string, errCh chan error) {
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
		tunName := tun.GetTunName("client", 1, 10+rand.Intn(191))
		tunIf, err := tun.ConnectToTun(ctx, tunName)
		if err != nil {
			errCh <- err
		}
		logger.Debug("connected to tun", zap.String("tun_name", tunName))

		brd := localnet.GetBrdFromIp(ctx, respStrings[1])
		if brd == "" {
			errCh <- errors.New("failed to get brd")
		}

		err = tun.SetTunUp(ctx, respStrings[1], brd, tunName)
		if err != nil {
			errCh <- err
		}
		logger.Debug("set tun up", zap.String("tun_name", tunName))

		go tun.HandleTunEvent(ctx, tunIf, conn, errCh)
		go tun.HandleConnTunEvent(ctx, tunIf, conn, errCh)
	}
}

func (c *Manager) makeRequest(ctx context.Context, msg string, cmdName string, addr string) (err error) {
	logger := ctxmeta.GetLogger(ctx)

	conn, err := net.Dial("tcp", addr+":"+c.Config.ServerPort)
	if err != nil {
		logger.Error("failed to connect to server", zap.Error(err))
		return err
	}

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

func (c *Manager) processCreateRequest(ctx context.Context) error {
	logger := ctxmeta.GetLogger(ctx)

	var name, password, addr string
	fmt.Println("Enter localnet name:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		logger.Error("got error while scanning create localnet name", zap.Error(err))
		return err
	}

	fmt.Println("Enter localnet password:")
	_, err = fmt.Scanf("%s", &password)
	if err != nil {
		logger.Error("got error while scanning create localnet password", zap.Error(err))
		return err
	}

	fmt.Println("Enter server addr:")
	_, err = fmt.Scanf("%s", &addr)
	if err != nil {
		logger.Error("got error while scanning addr", zap.Error(err))
		return err
	}

	msg := fmt.Sprintf("%s %s %s", commands.CreateCmd, name, password)
	return c.makeRequest(ctx, msg, commands.CreateCmd, addr)
}

func (c *Manager) processConnectRequest(ctx context.Context) error {
	logger := ctxmeta.GetLogger(ctx)

	var name, password, addr string
	fmt.Println("Enter localnet name:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		logger.Error("got error while scanning create localnet name", zap.Error(err))
		return err
	}

	fmt.Println("Enter localnet password:")
	_, err = fmt.Scanf("%s", &password)
	if err != nil {
		logger.Error("got error while scanning create localnet password", zap.Error(err))
		return err
	}

	fmt.Println("Enter server addr:")
	_, err = fmt.Scanf("%s", &addr)
	if err != nil {
		logger.Error("got error while scanning addr", zap.Error(err))
		return err
	}

	msg := fmt.Sprintf("%s %s %s", commands.ConnectCmd, name, password)
	return c.makeRequest(ctx, msg, commands.ConnectCmd, addr)
}

func (c *Manager) processCmd(ctx context.Context, cmd string) error {
	logger := ctxmeta.GetLogger(ctx)

	switch cmd {
	case commands.CreateCmd:
		return c.processCreateRequest(ctx)
	case commands.ConnectCmd:
		return c.processConnectRequest(ctx)
	default:
		logger.Error("undefined cmd")
		return errors.New("undefined cmd")
	}
}

func (c *Manager) RunClient(ctx context.Context) error {
	logger := ctxmeta.GetLogger(ctx)
	fmt.Println("Enter cmd:")

	var cmd string
	for {
		_, err := fmt.Scanf("%s", &cmd)
		if err != nil {
			logger.Error("got error while scanning cmd", zap.Error(err))
			return err
		}

		if err := c.processCmd(ctx, cmd); err != nil {
			return err
		}
	}
}
