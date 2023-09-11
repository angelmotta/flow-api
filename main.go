package main

import (
	"context"
	"fmt"
	"github.com/angelmotta/flow-api/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
)

func main() {
	c := config.Init()
	dbpool, err := pgxpool.New(context.Background(), c.GetPgDsn())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var greeting string
	err = dbpool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
}
