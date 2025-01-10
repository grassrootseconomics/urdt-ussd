package main

import (
	"context"
	"log"

	"git.grassecon.net/grassrootseconomics/visedriver/cmd/admin/commands"
)

func main() {
	ctx := context.Background()
	err := commands.Seed(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize a list of admins with error %s", err)
	}

}
