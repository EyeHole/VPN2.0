package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"regexp"
	"strings"

	"go.uber.org/zap"

	commands "VPN2.0/cmd"
	"VPN2.0/lib/ctxmeta"
)

func (s *Manager) RunServer(ctx context.Context, serverAddr string) error {
	logger := ctxmeta.GetLogger(ctx)

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

	_, err := conn.Write([]byte(result))
	if err != nil {
		logger.Error("failed to write to conn", zap.Error(err))
		return err
	}

	return nil
}


func (s *Manager) processCmd(ctx context.Context, cmd string, conn net.Conn) error {
	logger := ctxmeta.GetLogger(ctx)

	r := regexp.MustCompile("\\s+")
	replace := r.ReplaceAllString(cmd, " ")
	args := strings.Split(replace, " ")

	if len(args) < 1 {
		logger.Error("wrong cmd")
		return errors.New("wrong cmd")
	}

	resp := "wrong cmd"
	var cmdErr error
	switch args[0] {
	case commands.CreateCmd:
		logger.Debug("start processing create cmd")
		result, err := s.processNetworkCreationRequest(ctx, args, conn)
		resp = result
		cmdErr = err
	case commands.ConnectCmd:
		logger.Debug("start processing connect cmd")
		result, err := s.processConnectRequest(ctx, args, conn)
		resp = result
		cmdErr = err
	}

	if resp == "" {
		resp = "failed to process request"
	}

	err := sendResult(ctx, resp, conn)
	if err != nil {
		logger.Error("failed to send resp", zap.String("response", resp))
		return err
	}
	logger.Debug("sent resp", zap.String("response", resp))

	if cmdErr != nil {
		return err
	}

	return nil
}

func (s *Manager) handleClient(ctx context.Context, conn net.Conn, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	defer func() {
		err := conn.Close()
		if err != nil {
			logger.Error("failed to close conn", zap.Error(err))
			errCh <- err
		}
	}()

	clientReader := bufio.NewReader(conn)
	for {
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
}
