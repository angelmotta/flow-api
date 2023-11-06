package main

import (
	"context"
	"github.com/angelmotta/flow-api/api"
	"github.com/angelmotta/flow-api/database"
	"github.com/angelmotta/flow-api/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"net/http"
)

func main() {
	c := config.Init()

	dbpool, err := pgxpool.New(context.Background(), c.GetPgDsn())
	if err != nil {
		log.Fatalf("Error creating a new Pool connection from Postgres database: %v\n", err)
	}
	err = dbpool.Ping(context.Background())
	if err != nil {
		log.Fatalf("Unable to acquires a connection from the Pool and check it: %v\n", err)
	}
	defer dbpool.Close()
	log.Println("Successfully connected to Postgres database")

	// Create a store Object using the database pool
	store := database.NewPgStore(dbpool) // store Object implements the Store interface
	// Create a server by injecting the store as a dependency
	server := api.NewServer(store, c)

	// Chi router
	r := chi.NewRouter()
	r.Use(middleware.AllowContentType("application/json"))
	r.Use(middleware.RequestSize(server.Config.HttpMaxBodyBytes))
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		//AllowedOrigins: []string{"https://*", "http://*"},
		AllowedOrigins: []string{"http://localhost:1234"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Get("/api/v1/users", server.GetUsersHandler)
	r.Get("/api/v1/users/{email}", server.GetUserHandler)
	r.Post("/api/v1/users", server.CreateUserHandler)
	r.Post("/api/v1/users/signup", server.UserSignupHandler)
	r.Put("/api/v1/users/{id}", server.UpdateUserHandler)
	r.Delete("/api/v1/users/{id}", server.DeleteUserHandler)
	r.Post("/api/v1/auth/login", server.LoginHandler)
	log.Println("API server at port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
