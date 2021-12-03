package cache

import (
	"context"
	"go.uber.org/zap"
	"strconv"

	"github.com/gomodule/redigo/redis"

	"VPN2.0/lib/ctxmeta"
)

const (
	setValue = "set"
)

type Manager struct {
	redis redis.Conn
}

func NewCacheManager(ctx context.Context) (*Manager, error) {
	logger := ctxmeta.GetLogger(ctx)

	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		logger.Error("failed to set conn to redis server")
		return nil, err
	}

	return &Manager{
		redis: conn,
	}, nil
}

func (m *Manager) GetFirstAvailableClient(ctx context.Context, netID int, netCapacity int) (int, error) {
	logger := ctxmeta.GetLogger(ctx)
	clientID := -1
	for i := 1; i < netCapacity; i++ {
		isSet, err := redis.Bool(m.redis.Do("HEXISTS", strconv.Itoa(netID), strconv.Itoa(i)))
		if err != nil {
			logger.Error("failed to check clientID for existence", zap.Error(err))
			return -1, err
		}
		if !isSet {
			clientID = i
			break
		}
	}

	return clientID, nil
}

func (m *Manager) SetClient(ctx context.Context, netID int, clientID int) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := m.redis.Do("HSET", strconv.Itoa(netID), strconv.Itoa(clientID), setValue)
	if err != nil {
		logger.Error("failed to set clientID", zap.Error(err))
		return err
	}

	return nil
}
