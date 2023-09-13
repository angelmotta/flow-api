package main

import (
	"context"
	"github.com/angelmotta/flow-api/api"
	"github.com/angelmotta/flow-api/database"
	"github.com/angelmotta/flow-api/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
)

func main() {
	c := config.Init()

	dbpool, err := pgxpool.New(context.Background(), c.GetPgDsn())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	// Create a store dependency using the database pool
	store := database.NewStore(dbpool) // store implements the Store interface

	// Create a server by injecting the store as a dependency
	server := api.NewServer(store)

	// Using methods from server to perform operations
	// Read operation: retrieve user from database
	email := "angelmotta@gmail.com"
	u, err := server.GetUser(email)
	if err != nil {
		log.Printf("Error getting user: %v", err)
	}
	log.Println("User retrieved from database:")
	if u == nil {
		log.Println("User not found (user is available to be created)")
	} else {
		log.Println(u) // user value
	}

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
		// TODO: Check if error is for existing record or other error
		log.Printf("Error creating user: %v", err)
	}

	// Delete operation: delete user from database
	userId := 2
	err = server.DeleteUser(userId)
	if err != nil {
		log.Printf("Error deleting user: %v", err)
	}

	// Chi router
	r := chi.NewRouter()
	r.Get("/api/v1/users/{email}", server.GetUserHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
