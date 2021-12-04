package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"math"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"go.uber.org/zap"

	"VPN2.0/lib/cmd"
	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/localnet"
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

	serverConnName := localnet.GetConnName(network.ID, clientID)
	s.storage.AddTun(serverConnName, conn)

	ipAddr := fmt.Sprintf("%d.%d.%d.%d/%d", 10, network.ID, 0, clientID, network.Mask)

	respSuccess := fmt.Sprintf("%s %s", cmd.SuccessResponse, ipAddr)
	err = sendResult(ctx, respSuccess, conn)
	if err != nil {
		logger.Error("failed to send resp", zap.String("response", respSuccess))
		return err
	}
	logger.Debug("sent resp", zap.String("response", respSuccess))

	errCh := make(chan error, 1)
	go s.HandleConnEvent(ctx, conn, errCh)

	close(errCh)
	if err = <-errCh; err != nil {
		return err
	}

	return nil
}

func getNetworkCapacity(mask int) int {
	return int(math.Pow(2, float64(32-mask)) - 2)
}

func (s *Manager) HandleConnEvent(ctx context.Context, conn net.Conn, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	reader := bufio.NewReader(conn)
	for {
		var bufPool = make([]byte, 1500)
		n, err := reader.Read(bufPool)

		if err != nil {
			fmt.Println("read failed:", n, err)
		}

		validBuf := bufPool[:n]

		packet := gopacket.NewPacket(validBuf, layers.LayerTypeIPv4, gopacket.Default)
		ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
		if ipv4Layer == nil {
			logger.Error("ipv4 error")
			return
		}

		ipv4, _ := ipv4Layer.(*layers.IPv4)
		srcIP := ipv4.SrcIP.String()
		dstIP := ipv4.DstIP.String()

		fmt.Println("src: ", srcIP)
		fmt.Println("dest: ", dstIP)

		dstNetID, dstTunID := localnet.GetNetIdAndTapId(ctx, dstIP)
		dstConnName := fmt.Sprintf("conn%s_%s", dstNetID, dstTunID)

		dstConn, found := s.storage.GetTun(dstConnName)
		if !found {
			logger.Warn("failed to find conn", zap.Error(err))
			continue
		}

		n, err = dstConn.Write(packet.Data())
		if err != nil {
			logger.Error("failed to write to conn", zap.Error(err))
			errCh <- err
		}
	}
}
