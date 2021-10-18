package main

import (
	"fmt"

	"VPN2.0/server"
)

func main() {
	server.CreateServer()

	err := server.RunServer()
	if err != nil {
		fmt.Println(err)
	}
}
