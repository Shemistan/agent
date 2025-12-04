package main

import (
	"log"

	"github.com/Shemistan/agent/internal/app/agent"
)

func main() {
	if err := agent.Run(); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}
