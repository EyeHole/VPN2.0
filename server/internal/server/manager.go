package server

import (
	"context"

	"VPN2.0/lib/ctxmeta"
	"VPN2.0/server/internal/cache"
	"VPN2.0/server/internal/db"
)

type Manager struct {
	db    *db.Manager
	cache *cache.Manager
}

const (
	mask = 24
)

func CreateServer(ctx context.Context) (*Manager, error) {
	logger := ctxmeta.GetLogger(ctx)

	dbManager, err := db.NewDBManager(ctx)
	if err != nil {
		logger.Error("failed to create db manager")
		return nil, err
	}
	cacheManager, err := cache.NewCacheManager(ctx)
	if err != nil {
		logger.Error("failed to create cache manager")
		return nil, err
	}

	if err := dbManager.CreateNetworksTable(ctx); err != nil {
		return nil, err
	}
	logger.Debug("created networks table")

	return &Manager{
		db:    dbManager,
		cache: cacheManager,
	}, nil
}