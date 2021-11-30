package localnet

import (
	"context"
	"fmt"
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

func GetNetIdAndTunId(ctx context.Context, ipAddr string) (int, int) {
	logger := ctxmeta.GetLogger(ctx)

	octs := strings.Split(ipAddr, ".")
	if len(octs) < 4 {
		logger.Error("failed to split ip: " + ipAddr)
		return -1, -1
	}

	netID, err := strconv.Atoi(octs[1])
	if err != nil {
		logger.Error("failed to parse netID")
		return -1, -1
	}
	tunID, err := strconv.Atoi(octs[3])
	if err != nil {
		logger.Error("failed to parse tunID")
		return -1, -1
	}

	return netID, tunID
}
