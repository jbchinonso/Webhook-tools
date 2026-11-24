// Command seedadmin creates (or updates) an admin user from environment
// variables. Run explicitly, never on startup.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"

	"gateway/webhook/internal/commons/hashes"
	"gateway/webhook/internal/config"
	"gateway/webhook/internal/database"
	"gateway/webhook/internal/modules/users"
)

func main() {
	email := os.Getenv("ADMIN_EMAIL")
	password := os.Getenv("ADMIN_PASSWORD")
	if email == "" || password == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD must be set")
	}
	if len(password) < 12 {
		log.Fatal("ADMIN_PASSWORD must be at least 12 characters")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	db, err := database.New(cfg.DB, cfg.Log.Level)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	repo := users.NewRepository(db.Primary)

	if existing, err := repo.FindByEmail(ctx, email); err == nil {
		pwHash, err := hashes.HashPassword(password)
		if err != nil {
			log.Fatalf("hash: %v", err)
		}
		existing.Role = users.RoleSuperAdmin
		existing.Status = users.StatusActive
		existing.PasswordHash = pwHash
		if err := repo.Update(ctx, existing); err != nil {
			log.Fatalf("update: %v", err)
		}
		fmt.Printf("updated admin %s\n", email)
		return
	}

	pwHash, err := hashes.HashPassword(password)
	if err != nil {
		log.Fatalf("hash: %v", err)
	}
	u := &users.User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: pwHash,
		Role:         users.RoleSuperAdmin,
		Status:       users.StatusActive,
		FullName:     "Administrator",
	}
	if err := repo.Create(ctx, nil, u); err != nil {
		log.Fatalf("create: %v", err)
	}
	fmt.Printf("created admin %s\n", email)
}
