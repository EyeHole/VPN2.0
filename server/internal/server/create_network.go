package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strconv"

	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
)

func (s *Manager) processNetworkCreationRequest(ctx context.Context, args []string, conn net.Conn) (string, error) {
	resp := "network created"

	if len(args) < 3 {
		return "", errors.New("wrong args amount")
	}

	err := s.createNetwork(ctx, args[1], args[2])
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (s *Manager) createNetwork(ctx context.Context, name string, passwordHash string) error {
	netID, err := s.db.AddNetwork(ctx, name, passwordHash, mask)
	if err != nil {
		return err
	}

	return createBridge(ctx, strconv.Itoa(netID))
}

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
