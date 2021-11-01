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
									name CHAR(30) NOT NULL,
									password VARCHAR NOT NULL,
								)`
	insertNetworkQuery = `INSERT INTO networks (name, password) VALUES (?, ?)`
	getNetworkQuery    = `SELECT id, name, password FROM networks`
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

func (m *Manager) AddNetwork(ctx context.Context, name string, passwordHash string) error {
	logger := ctxmeta.GetLogger(ctx)

	_, err := m.db.Exec(insertNetworkQuery, name, passwordHash)
	if err != nil {
		logger.Error("failed to exec network insertion query", zap.Error(err))
		return err
	}

	return nil
}

func (m *Manager) GetNetwork(ctx context.Context, name string, passwordHash string) (*models.Network, error) {
	logger := ctxmeta.GetLogger(ctx)

	network := &models.Network{}
	err := m.db.QueryRow(getNetworkQuery, name, passwordHash).Scan(network.ID, network.Name, network.Password)
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
