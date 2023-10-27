package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/angelmotta/flow-api/database"
	"github.com/angelmotta/flow-api/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"io"
	"log"
	"net/http"
	"net/mail"
	"strings"
)

type Server struct {
	store  database.Store // store is a dependency defined as an interface
	Config *config.Config
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
func NewServer(store database.Store, c *config.Config) *Server {
	return &Server{
		store:  store,
		Config: c,
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
	tokensResponse, err := s.generateTokens(u)
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

type LoginRequest struct {
	Idp string `json:"idp"`
}

func (l *LoginRequest) Validate() error {
	if l.Idp == "" {
		return errors.New("missing required 'idp' field")
	}
	if l.Idp != "facebook" && l.Idp != "google" {
		return errors.New("invalid 'idp' value")
	}
	return nil
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

	// Get Body from request
	loginRequest := &LoginRequest{}
	err := s.DecodeJsonBody(w, r, loginRequest)
	if err != nil {
		log.Println("Error decoding json body -> ", err)
		sendJsonResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := loginRequest.Validate(); err != nil {
		log.Println("validate loginRequest error -> ", err)
		sendJsonResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	email, err := s.isValidExternalUserToken(token, loginRequest.Idp)
	if err != nil {
		log.Println("Invalid token")
		sendJsonResponse(w, err.Error(), http.StatusUnauthorized)
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
		log.Println("User not registered, please signup")
		errResponse := ErrorMessage{
			Message: "User not registered, please signup",
		}
		sendJsonResponse(w, errResponse, http.StatusNotFound)
		return
	}

	// Create App tokens for user: access token and refresh token
	tokensResponse, err := s.generateTokens(user)
	if err != nil {
		log.Println("Error trying to generate access token -> ", err)
		errResponse := ErrorMessage{
			Message: "Error while generating access to App Flow",
			Error:   "Error while generating access token for user",
		}
		sendJsonResponse(w, errResponse, http.StatusInternalServerError)
		return
	}

	// Create and send response message
	log.Println("User successfully logged in: sending response message")
	responseMessage := successfulUserAccessResponse{
		User:           user,
		tokensResponse: *tokensResponse,
	}
	sendJsonResponse(w, responseMessage, http.StatusOK)
}

type UserSignupRequest struct {
	Step     string                 `json:"step"`
	Idp      string                 `json:"idp"`
	UserInfo *UserInfoSignupRequest `json:"user_info"`
}

func (u *UserSignupRequest) Validate() error {
	if u.Idp == "" {
		return errors.New("missing required 'idp' field")
	}
	if u.Step != "1" && u.Step != "2" {
		return errors.New("invalid 'step' value")
	}
	if u.Step == "2" && u.UserInfo == nil {
		return errors.New("missing User Information in 'user_info' field")
	}
	if err := u.UserInfo.Validate(); err != nil {
		return err
	}
	return nil
}

type UserInfoSignupRequest struct {
	Dni               string `json:"dni"`
	Name              string `json:"name"`
	LastnameMain      string `json:"lastname_main"`
	LastnameSecondary string `json:"lastname_secondary"`
	Address           string `json:"address"`
}

func (u *UserInfoSignupRequest) Validate() error {
	if u.Dni == "" {
		return errors.New("user_info missing required 'dni' field")
	}
	if u.Name == "" {
		return errors.New("user_info missing required 'name' field")
	}
	if u.LastnameMain == "" {
		return errors.New("user_info missing required 'lastname_main' field")
	}
	if u.LastnameSecondary == "" {
		return errors.New("user_info missing required 'lastname_secondary' field")
	}
	if u.Address == "" {
		return errors.New("user_info missing required 'address' field")
	}
	return nil
}

// UserSignupHandler creates a user from a Signup request according to the step of onboarding
func (s *Server) UserSignupHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("UserSignupHandler")
	// Get token from Authorization header
	payloadAuthHeader := r.Header.Get("Authorization")
	if payloadAuthHeader == "" {
		log.Println("No token provided in Authorization header")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	payloadAuthHeader = strings.TrimSpace(payloadAuthHeader)
	token := strings.TrimPrefix(payloadAuthHeader, "Bearer ")

	// Get Body from request
	userSignupRequest := &UserSignupRequest{}
	err := s.DecodeJsonBody(w, r, userSignupRequest)
	if err != nil {
		sendJsonResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Println("Input received: ")
	log.Println(userSignupRequest)

	// Validate Input
	err = userSignupRequest.Validate()
	if err != nil {
		log.Println("validate userSignupRequest error ->", err)
		sendJsonResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Verify token according to idp specified in body request
	email, err := s.isValidExternalUserToken(token, userSignupRequest.Idp)
	if err != nil {
		log.Println("Invalid token")
		sendJsonResponse(w, err.Error(), http.StatusUnauthorized)
		return
	}
	log.Println("User email: ", email)

	// Handle Signup step flow
	if userSignupRequest.Step == "1" {
		// Verify if user is available and respond with error if not
		user, err := s.store.GetUser(email)
		if err != nil {
			log.Printf("Error getting user from database: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if user != nil {
			log.Println("User already registered")
			resMessage := struct{ Message string }{
				Message: "User already registered",
			}
			sendJsonResponse(w, resMessage, http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	} else if userSignupRequest.Step == "2" {
		// Create user in database
		user := &database.User{
			Email:             email,
			Role:              "customer",
			Dni:               userSignupRequest.UserInfo.Dni,
			Name:              userSignupRequest.UserInfo.Name,
			LastnameMain:      userSignupRequest.UserInfo.LastnameMain,
			LastnameSecondary: userSignupRequest.UserInfo.LastnameSecondary,
			Address:           userSignupRequest.UserInfo.Address,
		}
		err = s.store.CreateUser(user)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			errorHttpCode := getCreateUserHttpCode(err.Error())
			errorResponse := ErrorMessage{
				Message: err.Error(),
			}
			sendJsonResponse(w, errorResponse, errorHttpCode)
			return
		}
		// Create tokens for users: access token and refresh token
		tokensResponse, err := s.generateTokens(user)
		if err != nil {
			log.Println("Error trying to generate access token: ", err)
			errResponse := ErrorMessage{
				Message: "Error while generating access to App Flow",
				Error:   "Error while generating access token for user",
			}
			sendJsonResponse(w, errResponse, http.StatusInternalServerError)
		}
		// Create and send response message
		log.Println("User successfully logged in: sending response message")
		responseMessage := successfulUserAccessResponse{
			User:           user,
			tokensResponse: *tokensResponse,
		}
		sendJsonResponse(w, responseMessage, http.StatusOK)
	} else {
		log.Println("Invalid step value")
		sendJsonResponse(w, "Invalid step value", http.StatusBadRequest)
		return
	}

	sendJsonResponse(w, userSignupRequest, http.StatusOK)
}

func sendJsonResponse(w http.ResponseWriter, response interface{}, statusCode int) {
	responseJson, err := json.Marshal(response)
	if err != nil {
		// Marshal error (internal server error)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Set headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	// Write Json response
	_, err = w.Write(responseJson)
	if err != nil {
		log.Printf("Error sending response: %v", err.Error())
		return
	}
}
