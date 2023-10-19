package api

import (
	"github.com/angelmotta/flow-api/database"
	"github.com/golang-jwt/jwt/v5"
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
func generateAccessToken(userId string, expiresAt time.Time) (string, error) {
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

func generateRefreshToken(userId string, expiresAt time.Time) (string, error) {
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

func generateTokens(user *database.User) (*tokensResponse, error) {
	uId := strconv.Itoa(user.Id)
	accessTokenExpiresAt := time.Now().Add(10 * time.Minute)
	accessToken, err := generateAccessToken(uId, accessTokenExpiresAt)
	if err != nil {
		log.Fatal("Error generating access token", err)
		return nil, err
	}

	refreshTokenExpiresAt := time.Now().Add(24 * time.Hour * 7)
	refreshToken, err := generateRefreshToken(uId, refreshTokenExpiresAt)
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

func isValidExternalUserToken(token string) (string, bool) {
	// TODO: verify token from Google

	return "angel.motta@utec.com", true
}
