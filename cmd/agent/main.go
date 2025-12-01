package main

import (
	"log"

	"github.com/Shemistan/agent/internal/app/agent"
)

func main() {
	if err := agent.Run("app.toml"); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}
