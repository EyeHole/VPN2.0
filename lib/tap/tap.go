package tap

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	cmdchain "github.com/rainu/go-command-chain"
	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"go.uber.org/zap"

	lib "VPN2.0/lib/cmd"
	"VPN2.0/lib/ctxmeta"
	"VPN2.0/lib/localnet"
)

func getTapEther(tapName string) net.HardwareAddr {
	fmt.Println("tap name", tapName)

	addr := &bytes.Buffer{}
	err := cmdchain.Builder().
		Join("ip", "addr", "show", tapName).
		Join("grep", "ether").
		Finalize().WithOutput(addr).Run()
	if err != nil {
		fmt.Println("error", err)
	}

	addrWords := lib.GetWords(addr.String())

	if len(addrWords) < 3 {
		mac, _ := net.ParseMAC("")
		return mac
	}

	fmt.Println("ether: ", addrWords[2])

	mac, _ := net.ParseMAC(addrWords[2])
	return mac
}

func ConnectToTap(ctx context.Context, tapName string) (*water.Interface, error) {
	logger := ctxmeta.GetLogger(ctx)

	config := water.Config{
		DeviceType: water.TAP,
	}
	config.Name = tapName

	ifce, err := water.New(config)
	if err != nil {
		logger.Error("failed to connect to tap interface", zap.Error(err))
		return nil, err
	}
	return ifce, nil
}

func SetTapUp(ctx context.Context, addr string, brd string, tapName string) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := exec.Command("ip", "a", "add", addr, "dev", tapName, "broadcast", brd).Output()
	if err != nil {
		logger.Error("failed to add tap interface", zap.Error(err))
		return err
	}

	_, err = exec.Command("ip", "link", "set", "dev", tapName, "up").Output()
	if err != nil {
		logger.Error("failed to set tap interface up", zap.Error(err))
		return err
	}

	return nil
}

func GetTapName(serviceName string, netID int, clientID int) string {
	return fmt.Sprintf("%s_tap%d_%d", serviceName, netID, clientID)
}

func PrepareFrame(src string, dst string, size int, ctx context.Context) ethernet.Frame {
	var frame ethernet.Frame

	dstNetID, dstTapID := localnet.GetNetIdAndTapId(ctx, dst)
	dstTapName := fmt.Sprintf("server_tap%s_%s", dstNetID, dstTapID)
	dstEther := getTapEther(dstTapName)

	srcNetID, srcTapID := localnet.GetNetIdAndTapId(ctx, src)
	srcTapName := fmt.Sprintf("server_tap%s_%s", srcNetID, srcTapID)
	srcEther := getTapEther(srcTapName)

	frame.Prepare(dstEther, srcEther, ethernet.NotTagged, ethernet.IPv4, size)
	return frame
}

func HandleTapEvent(ctx context.Context, tapIf *water.Interface, conn net.Conn, errCh chan error) {
	logger := ctxmeta.GetLogger(ctx)

	var frame ethernet.Frame

	for {
		frame.Resize(1500)
		n, err := tapIf.Read(frame)
		if err != nil {
			logger.Error("failed to read from tap", zap.Error(err))
			errCh <- err
		}
		frame = frame[:n]

		/*fmt.Println("TAP SOURCE ", frame.Source())
		fmt.Println("TAP DESTINATION ", frame.Destination())
		fmt.Println("TAP ", frame.Payload())*/
		fmt.Println("TAP ", tapIf.Name(), "\n GOT ", frame)

		//msg := string(frame.Payload())
		//logger.Info("got in tap", zap.String("payload", msg))

		_, err = conn.Write(frame.Payload())
		if err != nil {
			logger.Error("failed to write to conn", zap.Error(err))
			errCh <- err
		}
	}
}

func HandleConnEvent(ctx context.Context, tapIf *water.Interface, conn net.Conn, errCh chan error) {
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

		frame := PrepareFrame(srcIP, dstIP, len(validBuf), ctx)

		copy(frame.Payload(), validBuf)
		fmt.Println("PAYLOAD ", frame.Payload())

		fmt.Println("SEND PACKET", frame)
		n, err = tapIf.Write(frame)
		fmt.Println("WROTE ", n)
		if err != nil {
			logger.Error("failed to write to tap", zap.Error(err))
			errCh <- err
		}
	}
}
