package main

import (
	"log"

	"config-analyzer/internal/transport/grpc"
)

func main() {
	if err := grpc.Run(); err != nil {
		log.Fatal(err)
	}
}
