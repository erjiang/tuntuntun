package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("server or client?")
		return
	}

	if os.Args[0] == "client" {
		client()
	} else {
		server()
	}
}
