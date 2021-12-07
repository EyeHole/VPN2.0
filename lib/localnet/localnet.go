package localnet

import (
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"

	"VPN2.0/lib/ctxmeta"
)

// FIXME: now works only for 24 mask
func GetBrdFromIp(ctx context.Context, ipAddr string) string {
	logger := ctxmeta.GetLogger(ctx)

	octs := strings.Split(ipAddr, ".")
	if len(octs) < 4 {
		logger.Error("failed to split ip")
		return ""
	}

	return fmt.Sprintf("%s.%s.%s.255", octs[0], octs[1], octs[2])
}

func GetIDFromIp(ctx context.Context, ipAddr string) (int, error) {
	logger := ctxmeta.GetLogger(ctx)

	octs := strings.Split(ipAddr, ".")
	if len(octs) < 4 {
		logger.Error("failed to split ip")
		return -1, errors.New("failed to split ip")
	}

	id, err:= strconv.Atoi(strings.Split(octs[3], "/")[0])
	if err != nil {
		logger.Error("failed to parse id", zap.Error(err))
		return -1, err
	}

	return id, nil
}

func GetNetIdAndTunId(ctx context.Context, ipAddr string) (string, string) {
	logger := ctxmeta.GetLogger(ctx)

	octs := strings.Split(ipAddr, ".")
	if len(octs) < 4 {
		logger.Error("failed to split ip: " + ipAddr)
		return "", ""
	}

	return octs[1], octs[3]
}

func GetConnName(netID int, clientID int) string {
	return fmt.Sprintf("conn%d_%d", netID, clientID)
}