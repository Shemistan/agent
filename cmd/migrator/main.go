package main

import (
	"log"
	"os"

	"github.com/Shemistan/agent/internal/app/migrator"
)

func main() {
	migrationDir := "migration"
	if len(os.Args) > 1 {
		migrationDir = os.Args[1]
	}

	if err := migrator.Run("app.toml", migrationDir); err != nil {
		log.Fatalf("Migrator failed: %v", err)
	}
}
