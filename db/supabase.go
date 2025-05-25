package db

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Println("OMG! THERE ARE NO ENVIROMENTAL VARIABLES! WHAT DO I DO NOW?")
	}
}

func GetEnvVar(key string) string {
	return os.Getenv(key)
}

func CreateClient() *supabase.Client {
	url := GetEnvVar("URL")
	key := GetEnvVar("ANON_API_KEY")
	if url == "" || key == "" {
		log.Fatal("OMG! THERE ARE NO ENVs FOR THE CLIENT! WHAT DO I DO NOW?")
	}
	client, err := supabase.NewClient(url, key, nil)
	if err != nil {
		log.Fatalf("ooppppsieee!!! failed to create supabase client: %v", err)
	}
	return client
}
