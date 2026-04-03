package utils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWT(userID int, username string) (string, error) {
	key, exist := os.LookupEnv("JWT_SECRET")
	if !exist {
		log.Fatalf("JWT_SECRET is not defined")
		return "", fmt.Errorf("JWT_SECRET is not defined")
	}
	claims := jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(15 * time.Minute).Unix(),
		"iss":      "Belwa Madho",
		"iat":      time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}

func ParseJWT(tokenString string) (jwt.MapClaims, error) {
	key, exist := os.LookupEnv("JWT_SECRET")
	if !exist {
		log.Fatalf("JWT_SECRET PROBLEM")
		return nil, fmt.Errorf("JWT_SECRET PROBLEM")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(key), nil
	})
	if err != nil {
		log.Println("JWT parse error:", err) // log.Println, NOT log.Fatal — don't crash the server
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
