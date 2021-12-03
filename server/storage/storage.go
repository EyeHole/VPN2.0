package storage

import "github.com/songgao/water"

var tuns map[string]*water.Interface

func SetStorage() {
	tuns = make(map[string]*water.Interface)
}

func AddTun(name string, tun *water.Interface) {
	tuns[name] = tun
}

func GetTun(name string) (*water.Interface, bool) {
	tun, err := tuns[name]
	return tun, err
}
