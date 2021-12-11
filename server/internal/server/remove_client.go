package server

import (
	"VPN2.0/server/internal/security"
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"

	"go.uber.org/zap"

	"VPN2.0/lib/cmd"
	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/localnet"
)

func (s *Manager) processLeaveRequest(ctx context.Context, args []string, conn net.Conn) error {
	logger := ctxmeta.GetLogger(ctx)
	respErr := "failed_to_disconnect"

	if len(args) < 4 {
		logger.Error("wrong amount of args")
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return errors.New("wrong args amount")
	}

	network, err := s.db.GetNetwork(ctx, args[1])
	if err != nil {
		if err.Error() == cmd.NoNetworkResponse {
			respErr = err.Error()
		}
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		if err.Error() == cmd.NoNetworkResponse {
			return nil
		}
		return err
	}

	if !security.CheckPasswordHash(args[2], network.Password) {
		logger.Error("wrong password")
		respErr = "incorrect_password"
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return nil
	}

	clientID, err := strconv.Atoi(args[3])
	if err != nil {
		logger.Error("failed to parse clientID", zap.Error(err))

		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}

		return err
	}

	err = s.cache.RemoveClient(ctx, network.ID, clientID)
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}
	respSuccess := fmt.Sprintf("%s", cmd.SuccessResponse)
	err = sendResult(ctx, respSuccess, conn)
	if err != nil {
		logger.Error("failed to send resp", zap.String("response", respSuccess))
		return err
	}

	s.storage.DelConn(localnet.GetConnName(network.ID, clientID))

	return nil
}
