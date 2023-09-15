package api

import (
	"github.com/angelmotta/flow-api/database"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

type Server struct {
	store database.Store // store is a dependency defined as an interface
}

func NewServer(store database.Store) *Server {
	return &Server{store}
}

func (s *Server) GetUser(email string) (*database.User, error) {
	log.Printf("Getting user: %v", email)
	result, err := s.store.GetUser(email)
	if err != nil {
		return nil, err
	}
	// Could perform some validation before returning the result
	return result, nil
}

func (s *Server) CreateUser(user *database.User) error {
	log.Printf("Creating user: %v", user.Email)
	err := s.store.CreateUser(user)
	if err != nil {
		return err
	}
	// Could perform some validation before returning the result
	log.Println("User successfully created")
	return nil
}

func (s *Server) DeleteUser(id int) error {
	log.Printf("Deleting User with ID: %v", id)
	err := s.store.DeleteUser(id)
	if err != nil {
		return err
	}
	log.Println("User successfully deleted")
	return nil
}

// GetUserHandler HTTP Handler
func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	// TODO: Validate input (email)
	user, err := s.store.GetUser(email)
	if err != nil {
		log.Printf("Error getting user from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	if user == nil {
		log.Println("User not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}
	_, err = w.Write([]byte("User found: " + user.Email + " DNI: " + user.Dni))
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}
