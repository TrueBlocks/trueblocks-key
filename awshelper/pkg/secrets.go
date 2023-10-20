package awshelper

import (
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
