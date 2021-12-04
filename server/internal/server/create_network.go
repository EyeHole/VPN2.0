package server

import (
	"VPN2.0/lib/ctxmeta"
	"context"
	"errors"
	"go.uber.org/zap"
	"net"
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

	_, err := s.db.AddNetwork(ctx, args[1], args[2], mask)

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
