package tun

import (
	"bufio"
	"context"
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"net"
	"os/exec"

	"VPN2.0/lib/ctxmeta"
	"github.com/songgao/water"
	"go.uber.org/zap"
)

func ConnectToTun(ctx context.Context, tapName string) (*water.Interface, error) {
	logger := ctxmeta.GetLogger(ctx)

	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = tapName

	ifce, err := water.New(config)
	if err != nil {
		logger.Error("failed to connect to tap interface", zap.Error(err))
		return nil, err
	}
	return ifce, nil
}

func SetTunUp(ctx context.Context, addr string, brd string, tunName string) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := exec.Command("ip", "a", "add", addr, "dev", tunName, "broadcast", brd).Output()
	if err != nil {
		logger.Error("failed to add tun interface", zap.Error(err))
		return err
	}

	_, err = exec.Command("ip", "link", "set", "dev", tunName, "up").Output()
	if err != nil {
		logger.Error("failed to set tun interface up", zap.Error(err))
		return err
	}

	return nil
}

func GetTunName(serviceName string, netID int, clientID int) string {
	return fmt.Sprintf("%s_tun%d_%d", serviceName, netID, clientID)
}

func HandleTunEvent(ctx context.Context, tunIf *water.Interface, conn net.Conn, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	buffer := make([]byte, 1500)

	for {
		n, err := tunIf.Read(buffer)
		if err != nil {
			logger.Error("failed to read from tun", zap.Error(err))
			errCh <- err
		}
		validBuf := buffer[:n]

		packet := gopacket.NewPacket(validBuf, layers.LayerTypeIPv4, gopacket.Default)
		ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
		if ipv4Layer == nil {
			logger.Error("ipv4 error")
			return
		}

		//ipv4, _ := ipv4Layer.(*layers.IPv4)
		//srcIP := ipv4.SrcIP.String()
		//dstIP := ipv4.DstIP.String()

		//logger.Info("got in tap", zap.String("payload", msg))

		//packet = append(packet, []byte("\n")[0])
		_, err = conn.Write(packet.Data())
		if err != nil {
			logger.Error("failed to write to conn", zap.Error(err))
			errCh <- err
		}
	}
}

func HandleConnTunEvent(ctx context.Context, tunIf *water.Interface, conn net.Conn, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(tunIf)

	for {
		/*buf, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Warn("connection was closed")
				return
			}
			logger.Error("got error while reading from conn", zap.Error(err))
			errCh <- err
			return
		}*/

		//logger.Debug("got in conn", zap.String("buffer", buf))
		var packet = make([]byte, 1500)
		n, err := reader.Read(packet)

		if err != nil {
			fmt.Println("read failed:", n, err)
		}

		packet = packet[:n]

		_, err = writer.Write(packet)
		if err != nil {
			logger.Error("failed to write to tun", zap.Error(err))
			errCh <- err
		}
	}
}
