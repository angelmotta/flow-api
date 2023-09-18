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

func (s *Server) getUser(email string) (*database.User, error) {
	log.Printf("Getting user: %v", email)
	result, err := s.store.GetUser(email)
	if err != nil {
		return nil, err
	}
	// Could perform some validation before returning the result
	return result, nil
}

func (s *Server) createUser(user *database.User) error {
	log.Printf("Creating user: %v", user.Email)
	err := s.store.CreateUser(user)
	if err != nil {
		return err
	}
	// Could perform some validation before returning the result
	log.Println("User successfully created")
	return nil
}

func (s *Server) deleteUser(id int) error {
	log.Printf("Deleting User with ID: %v", id)
	err := s.store.DeleteUser(id)
	if err != nil {
		return err
	}
	log.Println("User successfully deleted")
	return nil
}

// GetUserHandler HTTP Handler returns a specific user
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
	// TODO: Return user as JSON

	_, err = w.Write([]byte("User found: " + user.Email + " DNI: " + user.Dni))
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}

// CreateUserHandler HTTP Handler creates a user
func (s *Server) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateUserHandler")
	_, err := w.Write([]byte("CreateUserHandler"))
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}

// DeleteUserHandler HTTP Handler deletes a user
func (s *Server) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("DeleteUserHandler")
	_, err := w.Write([]byte("DeleteUserHandler"))
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}

// UpdateUserHandler HTTP Handler updates some user fields
func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("UpdateUserHandler")
	_, err := w.Write([]byte("UpdateUserHandler"))
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}

// GetUsersHandler HTTP Handler returns all users
func (s *Server) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("GetUsersHandler")
	_, err := w.Write([]byte("GetUsersHandler"))
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}
