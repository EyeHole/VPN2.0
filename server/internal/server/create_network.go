package server

import (
	"context"
	"errors"
	"net"
	"os/exec"

	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
)

func (s *Manager) processNetworkCreationRequest(ctx context.Context, args []string, conn net.Conn) error {
	logger := ctxmeta.GetLogger(ctx)
	respErr := "failed to create network"

	if len(args) < 3 {
		logger.Error("wrong amount of args")

		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return errors.New("wrong args amount")
	}

	err := s.createNetwork(ctx, args[1], args[2])
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}

	respSuccess := "network created successfully"
	errSend := sendResult(ctx, respSuccess, conn)
	if errSend != nil {
		logger.Error("failed to send resp", zap.String("response", respSuccess))
		return errSend
	}
	logger.Debug("sent resp", zap.String("response", respSuccess))

	return errSend
}

func (s *Manager) createNetwork(ctx context.Context, name string, passwordHash string) error {
	_, err := s.db.AddNetwork(ctx, name, passwordHash, mask)
	if err != nil {
		return err
	}

	return nil
}

func createBridge(ctx context.Context, netID int) error {
	logger := ctxmeta.GetLogger(ctx)

	bridgeName := getBridgeName(netID)

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
