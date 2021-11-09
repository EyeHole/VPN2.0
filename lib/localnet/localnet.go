package localnet

import (
	"context"
	"fmt"
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

func GetNetIdAndTapId(ctx context.Context, ipAddr string) (string, string) {
	logger := ctxmeta.GetLogger(ctx)

	octs := strings.Split(ipAddr, ".")
	if len(octs) < 4 {
		logger.Error("failed to split ip: " + ipAddr)
		return "", ""
	}

	return octs[1], octs[3]
}
