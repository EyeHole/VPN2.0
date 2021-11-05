package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"strings"

	"go.uber.org/zap"

	commands "VPN2.0/lib/cmd"
	"VPN2.0/lib/ctxmeta"
)

func (s *Manager) RunServer(ctx context.Context, serverAddr string) error {
	logger := ctxmeta.GetLogger(ctx)

	logger.Info(serverAddr)
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		logger.Error("got error while trying to listen", zap.Error(err))
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Warn("got error while trying to accept conn", zap.Error(err))
			continue
		}

		logger.Debug("someone connected")

		errCh := make(chan error, 1)
		go s.handleClient(ctx, conn, errCh)

		close(errCh)
		if err := <-errCh; err != nil {
			return err
		}
	}
}

func sendResult(ctx context.Context, result string, conn net.Conn) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := conn.Write([]byte(result + "\n"))
	if err != nil {
		logger.Error("failed to write to conn", zap.Error(err))
		return err
	}

	return nil
}

func (s *Manager) processCmd(ctx context.Context, cmd string, conn net.Conn) error {
	logger := ctxmeta.GetLogger(ctx)

	args := commands.GetWords(cmd)

	if len(args) < 1 {
		logger.Error("wrong cmd")
		return errors.New("wrong cmd")
	}

	switch args[0] {
	case commands.CreateCmd:
		logger.Debug("start processing create cmd")
		return s.processNetworkCreationRequest(ctx, args, conn)

	case commands.ConnectCmd:
		logger.Debug("start processing connect cmd")
		return s.processConnectRequest(ctx, args, conn)
	}

	resp := "wrong cmd"
	err := sendResult(ctx, resp, conn)
	if err != nil {
		logger.Error("failed to send resp", zap.String("response", resp))
		return err
	}

	logger.Debug("sent resp", zap.String("response", resp))

	return nil
}

func (s *Manager) handleClient(ctx context.Context, conn net.Conn, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	//defer func() {
	//	err := conn.Close()
	//	if err != nil {
	//		logger.Error("failed to close conn", zap.Error(err))
	//		errCh <- err
	//	}
	//}()

	clientReader := bufio.NewReader(conn)
	cmd, err := clientReader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			logger.Warn("client closed the connection by terminating the process")
			return
		}
		logger.Error("got error while reading from client", zap.Error(err))
		errCh <- err
		return
	}
	cmd = strings.TrimSpace(cmd)
	logger.Debug("got cmd", zap.String("cmd", cmd))

	err = s.processCmd(ctx, cmd, conn)
	if err != nil {
		errCh <- err
	}
}
