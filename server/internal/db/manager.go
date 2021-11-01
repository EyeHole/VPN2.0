package db

import (
	"context"
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"

	"VPN2.0/lib/ctxmeta"
	"VPN2.0/server/internal/db/models"
)

type Manager struct {
	db *sql.DB
}

const (
	createNetworksTableQuery = `CREATE TABLE IF NOT EXISTS networks (
									id INTEGER PRIMARY KEY, 
									created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
									updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
									net_name VARCHAR(30) NOT NULL,
									password VARCHAR NOT NULL,
    								mask INTEGER NOT NULL
								);`
	insertNetworkQuery = `INSERT INTO networks (net_name, password, mask) VALUES (?, ?, ?)`
	getNetworkQuery    = `SELECT id, net_name, password, mask FROM networks WHERE net_name = $1 AND password = $2`
)

func NewDBManager(ctx context.Context) (*Manager, error) {
	logger := ctxmeta.GetLogger(ctx)

	db, err := sql.Open("sqlite3", "file:./database.db?cache=shared")
	if err != nil {
		logger.Error("failed to open sqlite db", zap.Error(err))
		return nil, err
	}

	return &Manager{
		db: db,
	}, nil
}

func (m *Manager) CreateNetworksTable(ctx context.Context) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := m.db.Exec(createNetworksTableQuery)
	if err != nil {
		logger.Error("failed to exec networks table creation query", zap.Error(err))
		return err
	}

	return nil
}

func (m *Manager) AddNetwork(ctx context.Context, name string, passwordHash string, mask int32) (int, error) {
	logger := ctxmeta.GetLogger(ctx)

	result, err := m.db.Exec(insertNetworkQuery, name, passwordHash, mask)
	if err != nil {
		logger.Error("failed to exec network insertion query", zap.Error(err))
		return -1, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		logger.Error("failed to get inserted row id", zap.Error(err))
		return -1, err
	}

	logger.Debug("added network to db", zap.Int("id", int(id)))

	return int(id), nil
}

func (m *Manager) GetNetwork(ctx context.Context, name string, passwordHash string) (*models.Network, error) {
	logger := ctxmeta.GetLogger(ctx)

	network := &models.Network{}
	err := m.db.QueryRow(getNetworkQuery, name, passwordHash).Scan(&network.ID, &network.NetName, &network.Password, &network.Mask)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Error("no network row with such args")
			return nil, errors.New("row not found")
		}

		logger.Error("failed to get network from db", zap.Error(err))
		return nil, err
	}

	return network, nil
}
