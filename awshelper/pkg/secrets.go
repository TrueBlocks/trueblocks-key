package awshelper

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

var secretCache *secretcache.Cache

func FetchSecret(id string) (result string, err error) {
	if secretCache == nil {
		secretCache, err = secretcache.New()
		if err != nil {
			err = fmt.Errorf("awshelper/secrets: init secret cache: %w", err)
			return
		}
	}

	return secretCache.GetSecretString(id)
}

type UsernamePasswordSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *UsernamePasswordSecret) FromString(rawValue string) error {
	return json.Unmarshal([]byte(rawValue), u)
}

func FetchUsernamePasswordSecret(id string) (result *UsernamePasswordSecret, err error) {
	raw, err := FetchSecret(id)
	if err != nil {
		return
	}
	result = &UsernamePasswordSecret{}
	err = result.FromString(raw)
	return
}
