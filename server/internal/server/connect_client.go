package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"

	"VPN2.0/lib/ctxmeta"
)

func (s *Manager) processConnectRequest(ctx context.Context, args []string, conn net.Conn) (string, error) {
	logger := ctxmeta.GetLogger(ctx)
	resp := "connected successfully"

	if len(args) < 3 {
		return "", errors.New("wrong args amount")
	}

	network, err := s.db.GetNetwork(ctx, args[1], args[2])
	if err != nil {
		return "", err
	}

	clientID, err := s.cache.GetFirstAvailableClient(ctx, network.ID, getNetworkCapacity(network.Mask))
	if err != nil {
		return "", err
	}

	if clientID < 0 {
		logger.Error("failed to get clientID in network - network is full")
		resp = "failed to connect, network is full"
		return resp, errors.New("network is full")
	}

	err = s.cache.SetClient(ctx, network.ID, clientID)
	if err != nil {
		return "", err
	}

	serverTapName := fmt.Sprintf("server_tap%d_%d", network.ID, clientID)
	resp += serverTapName

	return resp, nil

}

func getNetworkCapacity(mask int) int {
	return int(math.Pow(2, float64(32-mask)) - 2)
}
