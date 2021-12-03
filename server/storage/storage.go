package storage

import (
	"sync"

	"github.com/songgao/water"
)

type Storage struct {
	Tuns map[string]*water.Interface
	Mu   *sync.Mutex
}

//var Tuns *Storage

/*func SetStorage() {
	storage := Storage{
		Tuns: map[string]*water.Interface{},
		Mu:   &sync.Mutex{},
	}
	Tuns = &storage
}*/

/*
func AddTun(name string, tun *water.Interface) {
	Tuns.Mu.Lock()
	Tuns.Tuns[name] = tun
	Tuns.Mu.Unlock()
}

func GetTun(name string) (*water.Interface, bool) {
	Tuns.Mu.Lock()
	tun, err := Tuns.Tuns[name]
	Tuns.Mu.Unlock()
	return tun, err
}*/
