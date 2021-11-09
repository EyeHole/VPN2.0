package tap

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os/exec"

	"github.com/songgao/packets/ethernet"
	"github.com/songgao/water"
	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
)

func getBridgeEther() (net.HardwareAddr, error) {
	addr, _ := exec.Command("ip -o link | grep b- | grep ether | awk '{ print $17 }'").Output()
	return net.ParseMAC(string(addr))

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

func SetTapUp(ctx context.Context, addr string, tapName string) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := exec.Command("ip", "a", "add", addr, "dev", tapName).Output()
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

		msg := string(frame.Payload())
		//logger.Info("got in tap", zap.String("payload", msg))

		_, err = conn.Write([]byte(msg + "\n"))
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
		/*buf, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				logger.Warn("connection was closed")
				return
			}
			logger.Error("got error while reading from conn", zap.Error(err))
			errCh <- err
			return
		}
		*/
		//logger.Debug("got in conn", zap.String("buffer", buf))

		var bufPool = make([]byte, 1500)
		n, err := reader.Read(bufPool)

		if err != nil {
			fmt.Println("read failed:", n, err)
		}

		validBuf := bufPool[:n]
		fmt.Println("CONNECTION ", validBuf)

		var frame ethernet.Frame
		frame.Resize(len(validBuf))

		ether, _ := getBridgeEther()
		copy(frame.Destination(), ether)
		copy(frame.Payload(), validBuf)
		fmt.Println("PAYLOAD ", frame.Payload())

		_, err = tapIf.Write(frame)
		if err != nil {
			logger.Error("failed to write to tap", zap.Error(err))
			errCh <- err
		}
	}
}
