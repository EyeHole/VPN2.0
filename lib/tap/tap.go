package tap

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/songgao/water"
	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
)

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
