package client

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/songgao/water"
	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
)

func (c *Manager) HandleTunEvent(ctx context.Context, tunIf *water.Interface, wg *sync.WaitGroup, conn net.Conn, errCh chan error) {
	defer wg.Done()

	logger := ctxmeta.GetLogger(ctx)

	buffer := make([]byte, 1500)

	for {
		n, err := tunIf.Read(buffer)
		if err != nil {
			if c.ID == IDUndefined {
				break
			}

			logger.Error("failed to read from tun", zap.Error(err), zap.Int("clientID", c.ID))
			errCh <- err
			return
		}
		validBuf := buffer[:n]

		packet := gopacket.NewPacket(validBuf, layers.LayerTypeIPv4, gopacket.Default)
		ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
		if ipv4Layer == nil {
			logger.Error("ipv4 error")
			errCh <- errors.New("ipv4 error")
			return
		}

		_, err = conn.Write(packet.Data())
		if err != nil {
			logger.Error("failed to write to conn", zap.Error(err))
			errCh <- err
			return
		}
	}
}

func (c *Manager) HandleConnTunEvent(ctx context.Context, tunIf *water.Interface, wg *sync.WaitGroup, conn net.Conn, errCh chan error) {
	defer wg.Done()

	logger := ctxmeta.GetLogger(ctx)

	reader := bufio.NewReader(conn)
	for {
		var bufPool = make([]byte, 1500)
		n, err := reader.Read(bufPool)

		if err != nil {
			if err == io.EOF {
				logger.Warn("connection was closed")
				errCh <- err
				return
			}

			if c.ID == IDUndefined {
				break
			}

			logger.Error("failed to read from conn", zap.Error(err))
			errCh <- err
			return
		}

		validBuf := bufPool[:n]

		packet := gopacket.NewPacket(validBuf, layers.LayerTypeIPv4, gopacket.Default)
		ipv4Layer := packet.Layer(layers.LayerTypeIPv4)
		if ipv4Layer == nil {
			logger.Error("ipv4 error")
			errCh <- errors.New("ipv4 error")
			return
		}

		ipv4, _ := ipv4Layer.(*layers.IPv4)
		srcIP := ipv4.SrcIP.String()
		dstIP := ipv4.DstIP.String()

		fmt.Println("src: ", srcIP)
		fmt.Println("dest: ", dstIP)

		_, err = tunIf.Write(packet.Data())
		if err != nil {
			logger.Error("failed to write to tun", zap.Error(err))
			errCh <- err
			return
		}
	}
}
