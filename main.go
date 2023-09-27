package main

import (
	"context"
	"github.com/angelmotta/flow-api/api"
	"github.com/angelmotta/flow-api/database"
	"github.com/angelmotta/flow-api/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	store := database.NewPgStore(dbpool) // store implements the Store interface
	// Create a server by injecting the store as a dependency
	server := api.NewServer(store)

	// Chi router
	log.Println("Starting server at port 8080")
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.RequestSize(server.MaxBodyBytes))
	r.Get("/api/v1/users", server.GetUsersHandler)
	r.Get("/api/v1/users/{email}", server.GetUserHandler)
	r.Post("/api/v1/users", server.CreateUserHandler)
	r.Put("/api/v1/users/{id}", server.UpdateUserHandler)
	r.Delete("/api/v1/users/{id}", server.DeleteUserHandler)
	log.Fatal(http.ListenAndServe(":8080", r))
}
