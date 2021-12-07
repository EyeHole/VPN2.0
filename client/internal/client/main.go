package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"go.uber.org/zap"

	"VPN2.0/client/internal/tun"
	commands "VPN2.0/lib/cmd"
	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/localnet"
)

func (c *Manager) processResp(ctx context.Context, conn net.Conn, cmdName string, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	clientReader := bufio.NewReader(conn)
	resp, err := clientReader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			logger.Warn("connection was closed")
			errCh <- err
			return
		}
		logger.Error("failed to read from conn", zap.Error(err))
		errCh <- err
		return
	}

	logger.Info("Got resp", zap.String("resp", resp))

	switch cmdName {
	case commands.ConnectCmd:
		respStrings := commands.GetWords(resp)
		if len(respStrings) < 2 {
			logger.Error("empty resp from server")
			errCh <- errors.New("empty resp")
			return
		}
		if respStrings[0] != commands.SuccessResponse {
			logger.Error("got error in server resp")
			errCh <- errors.New("error in resp")
			return
		}

		id, err := localnet.GetIDFromIp(ctx, respStrings[1])
		if err != nil {
			errCh <- err
			return
		}

		c.SetClientID(id)

		tunName := tun.GetTunName("client", 1, c.ID)
		tunIf, err := tun.ConnectToTun(ctx, tunName)
		if err != nil {
			errCh <- err
			return
		}
		logger.Debug("connected to tun", zap.String("tun_name", tunName))

		brd := localnet.GetBrdFromIp(ctx, respStrings[1])
		if brd == "" {
			errCh <- errors.New("failed to get brd")
			return
		}

		err = tun.SetTunUp(ctx, respStrings[1], brd, tunName)
		if err != nil {
			errCh <- err
			return
		}
		logger.Debug("set tun up", zap.String("tun_name", tunName))

		var wg sync.WaitGroup
		wg.Add(1)
		go c.HandleTunEvent(ctx, tunIf, &wg, conn, errCh)
		wg.Add(1)
		go c.HandleConnTunEvent(ctx, tunIf, &wg, conn, errCh)
		wg.Wait()
	case commands.LeaveCmd:
		respStrings := commands.GetWords(resp)
		if respStrings[0] != commands.SuccessResponse {
			logger.Error("got error in server resp")
			errCh <- errors.New("error in resp")
			return
		}

		tunName := tun.GetTunName("client", 1, c.ID)
		c.SetClientID(IDUndefined)

		err := tun.SetTunDown(ctx, tunName)
		if err != nil {
			errCh <- err
			return
		}
	case commands.DeleteCmd:
		respStrings := commands.GetWords(resp)
		if respStrings[0] != commands.SuccessResponse {
			logger.Error("got error in server resp")
			errCh <- errors.New("error in resp")
			return
		}

		tunName := tun.GetTunName("client", 1, c.ID)
		c.SetClientID(IDUndefined)

		err := tun.SetTunDown(ctx, tunName)
		if err != nil {
			errCh <- err
			return
		}
	}
}

func (c *Manager) makeRequest(ctx context.Context, msg string, cmdName string, addr string, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	conn, err := net.Dial("tcp", addr+":"+c.Config.ServerPort)
	if err != nil {
		logger.Error("failed to connect to server", zap.Error(err))
		errCh <- err
		return
	}

	go c.processResp(ctx, conn, cmdName, errCh)
	_, err = conn.Write([]byte(msg + "\n"))
	if err != nil {
		logger.Error("failed to write to conn", zap.Error(err))
		errCh <- err
		return
	}

	logger.Debug("sent cmd")
}

func (c *Manager) processCreateRequest(ctx context.Context, errCh chan error, inputMutex *bool) {
	logger := ctxmeta.GetLogger(ctx)

	var name, password, addr string
	fmt.Println("Enter localnet name:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		logger.Error("got error while scanning create localnet name", zap.Error(err))
		errCh <- err
		return
	}

	fmt.Println("Enter localnet password:")
	_, err = fmt.Scanf("%s", &password)
	if err != nil {
		logger.Error("got error while scanning create localnet password", zap.Error(err))
		errCh <- err
		return
	}

	fmt.Println("Enter server addr:")
	_, err = fmt.Scanf("%s", &addr)
	if err != nil {
		logger.Error("got error while scanning addr", zap.Error(err))
		errCh <- err
		return
	}

	msg := fmt.Sprintf("%s %s %s", commands.CreateCmd, name, password)

	*inputMutex = true

	c.makeRequest(ctx, msg, commands.CreateCmd, addr, errCh)
}

func (c *Manager) processConnectRequest(ctx context.Context, errCh chan error, inputMutex *bool) {
	logger := ctxmeta.GetLogger(ctx)

	var name, password, addr string
	fmt.Println("Enter localnet name:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		logger.Error("got error while scanning create localnet name", zap.Error(err))
		errCh <- err
		return
	}

	fmt.Println("Enter localnet password:")
	_, err = fmt.Scanf("%s", &password)
	if err != nil {
		logger.Error("got error while scanning create localnet password", zap.Error(err))
		errCh <- err
		return
	}

	fmt.Println("Enter server addr:")
	_, err = fmt.Scanf("%s", &addr)
	if err != nil {
		logger.Error("got error while scanning addr", zap.Error(err))
		errCh <- err
		return
	}

	msg := fmt.Sprintf("%s %s %s", commands.ConnectCmd, name, password)

	*inputMutex = true

	c.makeRequest(ctx, msg, commands.ConnectCmd, addr, errCh)
}

func (c *Manager) processLeaveRequest(ctx context.Context, errCh chan error, inputMutex *bool) {
	logger := ctxmeta.GetLogger(ctx)

	var name, password, addr string
	fmt.Println("Enter localnet name:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		logger.Error("got error while scanning create localnet name", zap.Error(err))
		errCh <- err
		return
	}

	fmt.Println("Enter localnet password:")
	_, err = fmt.Scanf("%s", &password)
	if err != nil {
		logger.Error("got error while scanning create localnet password", zap.Error(err))
		errCh <- err
		return
	}

	fmt.Println("Enter server addr:")
	_, err = fmt.Scanf("%s", &addr)
	if err != nil {
		logger.Error("got error while scanning addr", zap.Error(err))
		errCh <- err
		return
	}

	msg := fmt.Sprintf("%s %s %s %d", commands.LeaveCmd, name, password, c.ID)

	*inputMutex = true

	c.makeRequest(ctx, msg, commands.LeaveCmd, addr, errCh)
}

func (c *Manager) processDeleteRequest(ctx context.Context, errCh chan error, inputMutex *bool) {
	logger := ctxmeta.GetLogger(ctx)

	var name, password, addr string
	fmt.Println("Enter localnet name:")
	_, err := fmt.Scanf("%s", &name)
	if err != nil {
		logger.Error("got error while scanning create localnet name", zap.Error(err))
		errCh <- err
		return
	}

	fmt.Println("Enter localnet password:")
	_, err = fmt.Scanf("%s", &password)
	if err != nil {
		logger.Error("got error while scanning create localnet password", zap.Error(err))
		errCh <- err
		return
	}

	fmt.Println("Enter server addr:")
	_, err = fmt.Scanf("%s", &addr)
	if err != nil {
		logger.Error("got error while scanning addr", zap.Error(err))
		errCh <- err
		return
	}

	msg := fmt.Sprintf("%s %s %s", commands.DeleteCmd, name, password)

	*inputMutex = true

	c.makeRequest(ctx, msg, commands.DeleteCmd, addr, errCh)
}

func (c *Manager) processCmd(ctx context.Context, cmd string, errCh chan error, inputMutex *bool) {
	logger := ctxmeta.GetLogger(ctx)

	switch cmd {
	case commands.CreateCmd:
		c.processCreateRequest(ctx, errCh, inputMutex)
	case commands.ConnectCmd:
		c.processConnectRequest(ctx, errCh, inputMutex)
	case commands.LeaveCmd:
		c.processLeaveRequest(ctx, errCh, inputMutex)
	case commands.DeleteCmd:
		c.processDeleteRequest(ctx, errCh, inputMutex)
	default:
		logger.Error("undefined cmd")
		errCh <- errors.New("undefined cmd")
		return
	}
}

func (c *Manager) RunClient(ctx context.Context) error {
	logger := ctxmeta.GetLogger(ctx)

	var cmd string

	inputMutex := true
	errCh := make(chan error)
	var caughtErr error
	for caughtErr == nil {
		select {
		case errCheck := <-errCh:
			logger.Error("caught error!", zap.Error(errCheck))
			caughtErr = errCheck
		default:
			if inputMutex {
				fmt.Println("Enter cmd:")

				_, err := fmt.Scanf("%s", &cmd)
				if err != nil {
					logger.Error("got error while scanning cmd", zap.Error(err))
					return err
				}

				inputMutex = false
				go c.processCmd(ctx, cmd, errCh, &inputMutex)
			}
		}
	}
	close(errCh)

	return nil
}
