package server_tun

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/localnet"
)

func GetTunName(serviceName string, netID int, clientID int) string {
	return fmt.Sprintf("%s_tun%d_%d", serviceName, netID, clientID)
}

func HandleServerConnEvent(ctx context.Context, conn net.Conn, errCh chan error) {
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

		dstNetID, dstTunID := localnet.GetNetIdAndTapId(ctx, dstIP)
		dstTunName := fmt.Sprintf("server_tun%s_%s", dstNetID, dstTunID)

		file, err := os.Open("/dev/" + dstTunName)
		if err != nil {
			logger.Error("failed to open tun", zap.Error(err))
			errCh <- err
		}

		file.Write(validBuf)
		if err != nil {
			logger.Error("failed to open tun", zap.Error(err))
			errCh <- err
		}
	}
}
