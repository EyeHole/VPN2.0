package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"

	"VPN2.0/lib/localnet"

	"go.uber.org/zap"

	"VPN2.0/lib/cmd"
	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/tap"
)

func (s *Manager) processConnectRequest(ctx context.Context, args []string, conn net.Conn) error {
	logger := ctxmeta.GetLogger(ctx)
	respErr := "failed to connect"

	if len(args) < 3 {
		logger.Error("wrong amount of args")
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return errors.New("wrong args amount")
	}

	network, err := s.db.GetNetwork(ctx, args[1], args[2])
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}

	clientID, err := s.cache.GetFirstAvailableClient(ctx, network.ID, getNetworkCapacity(network.Mask))
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
	}

	if clientID < 0 {
		logger.Error("failed to get clientID in network - network is full")
		respErr = "failed to connect, network is full"

		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return errors.New("network is full")
	}

	err = s.cache.SetClient(ctx, network.ID, clientID)
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}

	serverTunName := tap.GetTunName("server", network.ID, clientID)
	tunAddr := fmt.Sprintf("%d.%d.%d.%d/%d", 10, network.ID, 0, clientID, network.Mask)

	tunIf, err := tap.ConnectToTun(ctx, serverTunName)
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}

	brd := localnet.GetBrdFromIp(ctx, tunAddr)
	if brd == "" {
		return errors.New("failed to get brd")
	}

	err = tap.SetTunUp(ctx, tunAddr, brd, serverTunName)
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}
	logger.Debug("Set tun up", zap.String("tun_name", serverTunName))

	//storage.AddTun(serverTunName, tunIf)

	respSuccess := fmt.Sprintf("%s %s", cmd.SuccessResponse, tunAddr)
	err = sendResult(ctx, respSuccess, conn)
	if err != nil {
		logger.Error("failed to send resp", zap.String("response", respSuccess))
		return err
	}
	logger.Debug("sent resp", zap.String("response", respSuccess))

	errCh := make(chan error, 1)
	go tap.HandleTunEvent(ctx, tunIf, conn, errCh)
	go tap.HandleConnEvent(ctx /* tunIf,*/, conn, errCh)

	close(errCh)
	if err = <-errCh; err != nil {
		return err
	}

	return nil
}

func getNetworkCapacity(mask int) int {
	return int(math.Pow(2, float64(32-mask)) - 2)
}
