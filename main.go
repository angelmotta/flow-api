package main

import (
	"context"
	"github.com/angelmotta/flow-api/api"
	"github.com/angelmotta/flow-api/database"
	"github.com/angelmotta/flow-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

func main() {
	c := config.Init()
	dbpool, err := pgxpool.New(context.Background(), c.GetPgDsn())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	// Create a store dependency using the database pool
	store := database.NewStore(dbpool) // store type implements store interface

	// Create a server by injecting the store as a dependency
	server := api.NewServer(store)

	// Use methods from the server type
	// Read operation: retrieve user from database
	u, err := server.GetUser(2)
	if err != nil {
		log.Fatalf("Error getting user: %v", err)
	}
	log.Println("User retrieved from database:")
	log.Println(u)

	// Write operation: create user in database
	newUser := database.User{
		Email:             "angelmotta@gmail.com",
		Role:              "customer",
		Dni:               "12345678",
		Name:              "Angel",
		LastnameMain:      "Motta",
		LastnameSecondary: "Paz",
		Address:           "Av. Los Incas 123",
	}
	err = server.CreateUser(&newUser)
	if err != nil {
		log.Println("Error creating user:", err)
	}
}
