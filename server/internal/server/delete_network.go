package server

import (
	"VPN2.0/lib/cmd"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"net"

	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/localnet"
)

func (s *Manager) processDeleteRequest(ctx context.Context, args []string, conn net.Conn) error {
	logger := ctxmeta.GetLogger(ctx)
	respErr := "failed to delete network"

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

	clientIDs, err := s.cache.GetAllClients(ctx, network.ID)
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}

	err = s.cache.RemoveNetwork(ctx, network.ID)
	if err != nil {
		errSend := sendResult(ctx, respErr, conn)
		if errSend != nil {
			logger.Error("failed to send resp", zap.String("response", respErr))
			return errSend
		}
		return err
	}

	err = s.db.DeleteNetwork(ctx, network.NetName, network.Password)

	respSuccess := fmt.Sprintf("%s", cmd.SuccessResponse)
	err = sendResult(ctx, respSuccess, conn)
	if err != nil {
		logger.Error("failed to send resp", zap.String("response", respSuccess))
		return err
	}

	for _, clientID := range clientIDs {
		connName := localnet.GetConnName(network.ID, clientID)
		// TODO: Send all clients in network "network_deleted" signal
		// doesn't work now because client can't process such messages from server

		//clientConn, _ := s.storage.GetConn(connName)
		//
		//msgDeleted := "network was deleted"
		//err = sendResult(ctx, msgDeleted, clientConn)
		//if err != nil {
		//	logger.Error("failed to send msg", zap.String("msg", msgDeleted))
		//	return err
		//}

		s.storage.DelConn(connName)
	}

	return nil
}
