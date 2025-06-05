package query

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"codejudger/db"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

type User struct {
	ID                string   `json:"id"`
	CreatedAt         string   `json:"created_at"`
	Bio               string   `json:"bio"`
	ProfilePicture    string   `json:"profile_picture"`
	Username          string   `json:"username"`
	FullName          string   `json:"full_name"`
	Slug              string   `json:"slug"`
	PrgLanguages      []string `json:"prg_languages"`
	GithubAccount     string   `json:"githubAccount"`
	DiscordAccount    string   `json:"discordAccount"`
	ShowLinkedGithub  bool     `json:"show_linked_github"`
	ShowLinkedDiscord bool     `json:"show_linked_discord"`
	ShowLinkedEmail   bool     `json:"show_linked_email"`
	ShowProfile       bool     `json:"show_profile"`
	APIKey            string   `json:"api_key"`
	Role              string   `json:"role"`
	Email             string   `json:"email"`
	Submissions       string   `json:"submissions"`
	CompletedDailies  string   `json:"completed_dailies"`
	JWT               string   `json:"jwt"`
}

func GetUserByJWT(jwtToken string) (*User, error) {
	_ = godotenv.Load()

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, errors.New("JWT_SECRET not set")
	}

	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid JWT")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims in token")
	}

	apiKey, ok := claims["sub"].(string)
	if !ok || apiKey == "" {
		return nil, errors.New("api_key not found in token claims")
	}

	client := db.CreateClient()

	rawData, _, err := client.
		From("users").
		Select("*", "", false).
		Or(fmt.Sprintf("api_key.eq.%s,id.eq.%s", apiKey, apiKey), "").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("error fetching user: %w", err)
	}

	var users []User
	if err := json.Unmarshal(rawData, &users); err != nil {
		return nil, errors.New("unable to parse user data")
	}
	if len(users) == 0 {
		return nil, errors.New("user not found")
	}

	return &users[0], nil
}
