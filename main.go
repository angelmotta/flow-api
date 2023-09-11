package main

import (
	"context"
	"fmt"
	"github.com/angelmotta/flow-api/database"
	"github.com/angelmotta/flow-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
)

type Server struct {
	Store database.Store // Store is a dependency defined as an interface
}

func main() {
	c := config.Init()
	dbpool, err := pgxpool.New(context.Background(), c.GetPgDsn())
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbpool.Close()

	// Create a store dependency with the database pool
	store := database.NewStore(dbpool)

	// Create a server by injecting the store as a dependency
	server := &Server{store}
	u, err := server.Store.GetUser(1)
	if err != nil {
		log.Fatalf("Error getting user: %v", err)
	}
	fmt.Println(u)

	//var greeting string
	//err = dbpool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	//if err != nil {
	//	fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
	//	os.Exit(1)
	//}

	//fmt.Println(greeting)
}
