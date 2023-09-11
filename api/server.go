package api

import (
	"errors"
	"github.com/angelmotta/flow-api/database"
	"github.com/jackc/pgx/v5/pgconn"
	"log"
)

type Server struct {
	store database.Store // store is a dependency defined as an interface
}

func NewServer(store database.Store) *Server {
	return &Server{store}
}

func (s *Server) GetUser(id int) (*database.User, error) {
	log.Println("Getting user...")
	result, err := s.store.GetUser(id)
	if err != nil {
		return nil, err
	}
	// Could perform some validation before returning the result
	return result, nil
}

func (s *Server) CreateUser(user *database.User) error {
	log.Println("Creating user...")
	err := s.store.CreateUser(user)
	var pgErr *pgconn.PgError
	if err != nil {
		log.Println("Error captured from database layer", err)
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return errors.New("user already exists")
			}
		}
		return errors.New("internal server error")
	}
	// Could perform some validation before returning the result
	return nil
}
