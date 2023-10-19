package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/angelmotta/flow-api/database"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"io"
	"log"
	"net/http"
	"net/mail"
	"strings"
)

type Server struct {
	store        database.Store // store is a dependency defined as an interface
	MaxBodyBytes int64          // The maximum number of bytes the server will read parsing the request body
}

type userCreateRequest struct {
	Email             string `json:"email"`
	Dni               string `json:"dni"`
	Name              string `json:"name"`
	LastnameMain      string `json:"lastname_main"`
	LastnameSecondary string `json:"lastname_secondary"`
	Address           string `json:"address"`
}

func (u *userCreateRequest) Bind(r *http.Request) error {
	if u.Email == "" {
		return errors.New("missing required 'email' field")
	}
	if u.Dni == "" {
		return errors.New("missing required 'dni' field")
	}
	if u.Name == "" {
		return errors.New("missing required 'name' field")
	}
	if u.LastnameMain == "" {
		return errors.New("missing required 'lastname_main' field")
	}
	if u.LastnameSecondary == "" {
		return errors.New("missing required 'lastname_secondary' field")
	}
	if u.Address == "" {
		return errors.New("missing required 'address' field")
	}
	return nil
}

// validateRequestFields validates the userCreateRequest fields of the request
func (u *userCreateRequest) validateRequestFields() error {
	if u.Email == "" {
		return errors.New("missing required 'email' field")
	}
	if !isValidEmail(u.Email) {
		return errors.New("invalid email")
	}
	if u.Dni == "" {
		return errors.New("missing required 'dni' field")
	}
	if u.Name == "" {
		return errors.New("missing required 'name' field")
	}
	if u.LastnameMain == "" {
		return errors.New("missing required 'lastname_main' field")
	}
	if u.LastnameSecondary == "" {
		return errors.New("missing required 'lastname_secondary' field")
	}
	if u.Address == "" {
		return errors.New("missing required 'address' field")
	}
	// We can perform more validations here
	return nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// NewServer receive an Interface Store and creates a new API Server Object
func NewServer(store database.Store) *Server {
	return &Server{
		store,
		1024 * 1024, // 1MB
	}
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

// DecodeJsonBody decodes the request body into the provided interface
func (s *Server) DecodeJsonBody(w http.ResponseWriter, r *http.Request, payload interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	err := dec.Decode(&payload)
	if err != nil {
		log.Printf("Got an error decoding json request")
		log.Println(err)
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			return &ErrResponse{Err: err, HTTPStatusCode: http.StatusBadRequest, StatusText: msg}

		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			return &ErrResponse{Err: err, HTTPStatusCode: http.StatusBadRequest, StatusText: msg}

		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			return &ErrResponse{Err: err, HTTPStatusCode: http.StatusBadRequest, StatusText: msg}

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			fieldName = strings.Trim(fieldName, "\"")
			log.Println("Unknown field: ", fieldName)
			msg := fmt.Sprintf("Unknown field '%s' in request", fieldName)
			return &ErrResponse{Err: err, HTTPStatusCode: http.StatusBadRequest, StatusText: msg}

		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return &ErrResponse{Err: err, HTTPStatusCode: http.StatusBadRequest, StatusText: msg}

		default:
			log.Println("Unknown error decoding json: ", err)
			msg := "Error reading and verifying request"
			return &ErrResponse{Err: err, HTTPStatusCode: http.StatusInternalServerError, StatusText: msg, ErrorText: err.Error()}
		}
	}
	return nil
}

// GetUserHandler HTTP Handler returns a specific user
func (s *Server) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	email := chi.URLParam(r, "email")
	if !isValidEmail(email) {
		log.Println("Invalid email")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO: verify token from Authorization header

	// Get user from database
	user, err := s.store.GetUser(email)
	if err != nil {
		log.Printf("Error getting user from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Println("User not found")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respJson, err := json.Marshal(user)
	if err != nil {
		log.Printf("Error marshalling user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//_, err = w.Write([]byte("User found: " + user.Email + " DNI: " + user.Dni))
	_, err = w.Write(respJson)
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}

type successfulUserAccessResponse struct {
	*database.User `json:"user_info"`
	tokensResponse `json:"tokens"`
}

// CreateUserHandler HTTP Handler creates a user from a Signup request
func (s *Server) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("CreateUserHandler")
	// Creating requestUserCreate 'Object' based on http request
	uCreateRequest := &userCreateRequest{}

	// Bind request body to userCreateRequest struct using custom bind function
	err := s.DecodeJsonBody(w, r, uCreateRequest)
	if err != nil {
		err := render.Render(w, r, err.(*ErrResponse)) // type assertion to access the interface value's underlying concrete value.
		if err != nil {
			log.Println("Error trying to render error: ", err)
			return
		}
		return
	}

	err = uCreateRequest.validateRequestFields()
	if err != nil {
		err := render.Render(w, r, ErrInvalidRequest(err))
		if err != nil {
			log.Println("Error trying to render error: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	log.Printf("valid request received: %v", uCreateRequest)

	// Create requestCreateUser Object based on HTTP request
	u := &database.User{
		Email:             uCreateRequest.Email,
		Role:              "customer",
		Dni:               uCreateRequest.Dni,
		Name:              uCreateRequest.Name,
		LastnameMain:      uCreateRequest.LastnameMain,
		LastnameSecondary: uCreateRequest.LastnameSecondary,
		Address:           uCreateRequest.Address,
	}

	// Create user record in database
	err = s.store.CreateUser(u)
	if err != nil {
		log.Printf("Error creating user: %v", err)
		httpStatusCode := getCreateUserHttpCode(err.Error())
		errorResponse := &ErrResponse{Err: err, HTTPStatusCode: httpStatusCode, StatusText: err.Error()}
		err := render.Render(w, r, errorResponse)
		if err != nil { // should never happen
			log.Println("Error trying to render error response: ", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		// Stop handler execution
		return
	}

	// Create tokens for users: access token and refresh token
	tokensResponse, err := generateTokens(u)
	if err != nil {
		log.Println("Error trying to generate access token: ", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Create response message
	responseMessage := successfulUserAccessResponse{
		User:           u,
		tokensResponse: *tokensResponse,
	}
	jsonResp, err := json.Marshal(responseMessage)
	if err != nil {
		log.Println("Error marshalling success response: ", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}

func getCreateUserHttpCode(errMsg string) int {
	if strings.Contains(errMsg, "A user already exists") {
		return http.StatusConflict
	}
	return http.StatusInternalServerError
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

// UpdateUserHandler HTTP Handler updates user fields as Bank Account
func (s *Server) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: UpdateUserHandler")
	_, err := w.Write([]byte("UpdateUserHandler"))
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}

// GetUsersHandler HTTP Handler returns a list of all users
func (s *Server) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("TODO: GetUsersHandler")
	_, err := w.Write([]byte("GetUsersHandler"))
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("LoginHandler")
	// Get token from Authorization header
	payloadAuthHeader := r.Header.Get("Authorization")
	if payloadAuthHeader == "" {
		log.Println("No token provided in Authorization header")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	payloadAuthHeader = strings.TrimSpace(payloadAuthHeader)
	token := strings.TrimPrefix(payloadAuthHeader, "Bearer ")

	// TODO: verify token from Authorization header and obtain email from payload
	email, ok := isValidExternalUserToken(token)
	if !ok {
		log.Println("Invalid token")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get user from database
	user, err := s.store.GetUser(email)
	if err != nil {
		log.Printf("Error getting user from database: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if user == nil {
		log.Println("User not found, please signup first")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// Create App tokens for user: access token and refresh token
	tokensResponse, err := generateTokens(user)
	if err != nil {
		log.Println("Error trying to generate access token: ", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	// Create response message
	log.Println("User successfully logged in: sending response message")
	responseMessage := successfulUserAccessResponse{
		User:           user,
		tokensResponse: *tokensResponse,
	}
	jsonResp, err := json.Marshal(responseMessage)
	if err != nil {
		log.Println("Error marshalling success response: ", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Println("Error writing http response: ", err)
		return
	}
}
