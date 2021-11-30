package server

import (
	"VPN2.0/lib/cmd"
	"VPN2.0/lib/localnet"
	"VPN2.0/lib/tun"
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"math"
	"net"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

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
	go s.HandleServerConnEvent(ctx, conn, network.ID, errCh)

	close(errCh)
	if err = <-errCh; err != nil {
		return err
	}

	return nil
}

func getNetworkCapacity(mask int) int {
	return int(math.Pow(2, float64(32-mask)) - 2)
}

func addTapToBridge(ctx context.Context, tapName string, bridgeName string) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := exec.Command("ip", "link", "set", tapName, "master", bridgeName).Output()
	if err != nil {
		logger.Error("failed to add tap to bridge", zap.Error(err))
		return err
	}

	return nil
}

func (s *Manager) HandleServerConnEvent(ctx context.Context, conn net.Conn, netID int, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	reader := bufio.NewReader(conn)
	for {
		var bufPool = make([]byte, 1500)
		n, err := reader.Read(bufPool)

		if err != nil {
			fmt.Println("read failed:", n, err)
		}

		validBuf := bufPool[:n]
		fmt.Println("CONNECTION ", validBuf)

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

		dstNetID, dstTunID := localnet.GetNetIdAndTunId(ctx, dstIP)
		goodDest, err := s.isInNetwork(ctx, netID, dstNetID, dstTunID)
		if err != nil{
			continue
		}
		if !goodDest {
			logger.Debug("caught bad packet")
			continue
		}

		dstTunName := fmt.Sprintf("server_tun%d_%d", dstNetID, dstTunID)

		// doesn't work
		tunFile, err := getTunFile(ctx, dstTunName)
		if err != nil {
			logger.Error("failed to get tun file", zap.Error(err))
			return
		}


		_, err = tunFile.Write(validBuf)
		if err != nil {
			logger.Error("failed to write to tun", zap.Error(err))
			errCh <- err
		}

		if err != nil {
			logger.Error("failed to open tun", zap.Error(err))
			errCh <- err
		}
	}
}

func getTunFile(ctx context.Context, tunName string) (*os.File, error) {
	logger := ctxmeta.GetLogger(ctx)

	file, err := os.Open("/dev/net/tun")
	if err != nil {
		logger.Error("failed to open tun file", zap.Error(err))
		return nil, err
	}

	ifr := make([]byte, 18)
	copy(ifr, tunName)
	ifr[16] = 0x01
	ifr[17] = 0x10

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(file.Fd()),
		uintptr(0x400454ca), uintptr(unsafe.Pointer(&ifr[0])))
	if errno != 0 {
		logger.Error("failed to open tun with ioctl", zap.String("errno", errno.Error()))
		return nil, errors.New(errno.Error())
	}

	return file, nil
}

func (s *Manager) isInNetwork(ctx context.Context, currentNetID int, dstNetID int, dstTunID int) (bool, error) {
	if currentNetID != dstNetID {
		return false, nil
	}

	tunExist, err := s.cache.ClientExist(ctx, currentNetID, dstTunID)
	if err != nil {
		return false, err
	}

	return tunExist, err
}
