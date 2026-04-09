package main

import (
	"log"

	"config-analyzer/internal/transport/rest"
)

func main() {
	if err := rest.Run(); err != nil {
		log.Fatal(err)
	}
}
