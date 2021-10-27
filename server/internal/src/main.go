package src

import (
	"VPN2.0/lib/ctxmeta"
	"bufio"
	"context"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net"
	"os/exec"
	"strings"

	"github.com/google/uuid"

	commands "VPN2.0/cmd"
)

func createBridge(ctx context.Context, networkID string) error {
	logger := ctxmeta.GetLogger(ctx)

	bridgeName := fmt.Sprintf("b-%s", networkID)

	_, err := exec.Command("ip", "link", "add", "name", bridgeName, "type", "bridge").Output()
	if err != nil {
		logger.Error("failed to exec bridge creation cmd", zap.Error(err))
		return err
	}
	logger.Debug("bridge created!", zap.String("bridge_name", bridgeName))

	_, err = exec.Command("ip", "link", "set", bridgeName, "up").Output()
	if err != nil {
		logger.Error("failed to exec bridge set up cmd", zap.Error(err))
		return err
	}
	logger.Debug("bridge is up!", zap.String("bridge_name", bridgeName))

	return nil
}

func RunServer(ctx context.Context, serverAddr string) error {
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
		go handleClient(ctx, conn, errCh)

		close(errCh)
		if err := <-errCh; err != nil {
			return err
		}
	}
}

func createNetwork(ctx context.Context, conn net.Conn) error {
	logger := ctxmeta.GetLogger(ctx)

	netID := uuid.New().String()[:5]
	err := createBridge(ctx, netID)

	resp := "network created"
	if err != nil {
		resp = "failed to process request"
	}

	_, err = conn.Write([]byte(resp))
	if err != nil {
		logger.Error("failed to write to conn", zap.Error(err))
	}

	return err
}

func processCmd(ctx context.Context, cmd string, conn net.Conn) error {
	switch cmd {
	case commands.CreateCmd:
		return createNetwork(ctx, conn)
	}

	return nil
}

func handleClient(ctx context.Context, conn net.Conn, errCh chan error) {
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

		err = processCmd(ctx, cmd, conn)
		if err != nil {
			errCh <- err
		}
	}
}

func CreateServer(ctx context.Context) {
}
