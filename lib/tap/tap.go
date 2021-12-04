package tap

import (
	"VPN2.0/lib/localnet"
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
	"VPN2.0/server/storage"
)

var stor = storage.Storage{
	Tuns: map[string]*water.Interface{},
	Mu:   &sync.Mutex{},
}

func ConnectToTun(ctx context.Context, tunName string) (*water.Interface, error) {
	logger := ctxmeta.GetLogger(ctx)

	config := water.Config{
		DeviceType: water.TUN,
	}
	config.Name = tunName

	ifce, err := water.New(config)
	if err != nil {
		logger.Error("failed to connect to tap interface", zap.Error(err))
		return nil, err
	}

	stor.Mu.Lock()
	stor.Tuns[tunName] = ifce
	stor.Mu.Unlock()

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

		ipv4, _ := ipv4Layer.(*layers.IPv4)
		srcIP := ipv4.SrcIP.String()
		dstIP := ipv4.DstIP.String()
		fmt.Println("in conn, src: ", srcIP, "dest: ", dstIP)
		//logger.Info("got in tap", zap.String("payload", msg))

		_, err = conn.Write(packet.Data())
		if err != nil {
			logger.Error("failed to write to conn", zap.Error(err))
			errCh <- err
		}
	}
}

func HandleConnEvent(ctx context.Context /*, tun *water.Interface*/, conn net.Conn, errCh chan error) {
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

		stor.Mu.Lock()
		tun, found := stor.Tuns[dstTunName]
		stor.Mu.Unlock()
		if !found {
			logger.Warn("failed to find tun", zap.Error(err))
			continue
		}

		n, err = tun.Write(packet.Data())
		fmt.Println("WROTE ", n)
		if err != nil {
			logger.Error("failed to write to tun", zap.Error(err))
			errCh <- err
		}
	}
}
