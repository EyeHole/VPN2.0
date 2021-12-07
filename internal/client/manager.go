package client

import "VPN2.0/client/internal/config"

const (
	IDUndefined = -1
)
type Manager struct {
	Config *config.Config
	ID     int
}

func (c *Manager) SetClientID(id int) {
	c.ID = id
}
