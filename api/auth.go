package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/angelmotta/flow-api/database"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/api/idtoken"
	"log"
	"strconv"
	"time"
)

type MyCustomClaims struct {
	jwt.RegisteredClaims
}

type tokensResponse struct {
	AccessToken           string    `json:"access_token"`
	ExpiresAt             time.Time `json:"expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`
}

// generateAccessToken generates a signed access token
func (s *Server) generateAccessToken(userId string, expiresAt time.Time) (string, error) {
	mySigningKey := []byte("my_dirty_secret")

	// Create claims with multiple fields populated
	claims := MyCustomClaims{
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "Flow App",
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(mySigningKey)
	if err != nil {
		log.Fatal("Error signing jwt", err)
		return "", err
	}
	return signedToken, nil
}

func (s *Server) generateRefreshToken(userId string, expiresAt time.Time) (string, error) {
	mySigningKey := []byte("my_dirty_secret")
	claims := MyCustomClaims{
		jwt.RegisteredClaims{
			// A usual scenario is to set the expiration time relative to the current time
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "Flow App",
			Subject:   userId,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(mySigningKey)
	if err != nil {
		log.Fatal("Error signing jwt", err)
		return "", err
	}
	return signedToken, nil
}

func (s *Server) generateTokens(user *database.User) (*tokensResponse, error) {
	uId := strconv.Itoa(user.Id)
	accessTokenExpiresAt := time.Now().Add(10 * time.Minute)
	accessToken, err := s.generateAccessToken(uId, accessTokenExpiresAt)
	if err != nil {
		log.Fatal("Error generating access token", err)
		return nil, err
	}

	refreshTokenExpiresAt := time.Now().Add(24 * time.Hour * 7)
	refreshToken, err := s.generateRefreshToken(uId, refreshTokenExpiresAt)
	if err != nil {
		log.Fatal("Error generating refresh token", err)
		return nil, err
	}

	r := &tokensResponse{
		AccessToken:           accessToken,
		ExpiresAt:             accessTokenExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}

	return r, nil
}

func (s *Server) verifyGTokenId(token string) (string, error) {
	// Verify the ID token, including the expiry, signature, issuer, and audience.
	tokenPayload, err := idtoken.Validate(context.Background(), token, s.Config.GOauthClientId)
	if err != nil {
		log.Printf("idtoken.Validate() error -> %v", err)
		return "", err
	}

	// Valid Token, you can use the token to get user information.
	fmt.Println("Token successfully verified.")
	email := tokenPayload.Claims["email"].(string)
	return email, nil
}

func (s *Server) isValidExternalUserToken(token, idp string) (string, error) {
	email := ""
	if idp == "google" {
		e, err := s.verifyGTokenId(token)
		if err != nil {
			return "", err
		}
		email = e
	} else if idp == "facebook" {
		log.Println("facebook not implemented yet")
		return "", errors.New("facebook not implemented yet")
	} else {
		log.Println("Invalid idp:", idp)
		return "", errors.New("invalid idp")
	}
	return email, nil
}
