package awshelper

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

var secretCache *secretcache.Cache
var TestSecret = &UsernamePasswordSecret{
	Username: "test",
	Password: "1234",
}

type UsernamePasswordSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func FetchSecret(id string) (result string, err error) {
	if secretCache == nil {
		secretCache, err = secretcache.New()
		if err != nil {
			err = fmt.Errorf("awshelper/secrets: init secret cache: %w", err)
			return
		}
	}

	log.Println("Fetching secret with ID:", id)

	return secretCache.GetSecretString(id)
}

func (u *UsernamePasswordSecret) FromString(rawValue string) error {
	return json.Unmarshal([]byte(rawValue), u)
}

func FetchUsernamePasswordSecret(id string) (result *UsernamePasswordSecret, err error) {
	if os.Getenv("AWS_SAM_LOCAL") == "true" {
		log.Println("Returning TEST secret")
		return TestSecret, nil
	}

	raw, err := FetchSecret(id)
	if err != nil {
		return
	}
	result = &UsernamePasswordSecret{}
	err = result.FromString(raw)
	return
}
