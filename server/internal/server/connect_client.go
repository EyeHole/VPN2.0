package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net"

	"VPN2.0/lib/cmd"
	"VPN2.0/lib/localnet"
	"VPN2.0/lib/tun"
	"VPN2.0/server/storage"

	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/server_tun"
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

	serverTapName := server_tun.GetTunName("server", network.ID, clientID)
	tapAddr := fmt.Sprintf("%d.%d.%d.%d/%d", 10, network.ID, 0, clientID, network.Mask)

	tapIf, err := tun.ConnectToTun(ctx, serverTapName)
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}

	storage.Tuns[serverTapName] = tapIf

	brd := localnet.GetBrdFromIp(ctx, tapAddr)
	if brd == "" {
		return errors.New("failed to get brd")
	}

	err = tun.SetTunUp(ctx, tapAddr, brd, serverTapName)
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}
	logger.Debug("Set tap up", zap.String("tap_name", serverTapName))

	/*err = addTapToBridge(ctx, serverTapName, getBridgeName(network.ID))
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}

	logger.Debug("Added tap to bridge", zap.String("tap_name", serverTapName), zap.String("bridge_name", getBridgeName(network.ID)))
	*/
	respSuccess := fmt.Sprintf("%s %s", cmd.SuccessResponse, tapAddr)
	err = sendResult(ctx, respSuccess, conn)
	if err != nil {
		logger.Error("failed to send resp", zap.String("response", respSuccess))
		return err
	}
	logger.Debug("sent resp", zap.String("response", respSuccess))

	errCh := make(chan error, 1)
	go tun.HandleTunEvent(ctx, tapIf, conn, errCh)
	go server_tun.HandleServerConnEvent(ctx, conn, errCh)

	close(errCh)
	if err = <-errCh; err != nil {
		return err
	}

	return nil
}

func getNetworkCapacity(mask int) int {
	return int(math.Pow(2, float64(32-mask)) - 2)
}

/*func addTapToBridge(ctx context.Context, tapName string, bridgeName string) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := exec.Command("ip", "link", "set", tapName, "master", bridgeName).Output()
	if err != nil {
		logger.Error("failed to add tap to bridge", zap.Error(err))
		return err
	}

	return nil
}*/
