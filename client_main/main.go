package main

import (
	"fmt"

	"VPN2.0/client"
)

func main() {
	err := client.RunClient()
	if err != nil {
		fmt.Println(err)
	}
}
