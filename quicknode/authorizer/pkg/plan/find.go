package plan

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
)

// This is our QN plan slug to API key value cache
var planSlugToApiKey = make(map[string]string)

// FindBySlug fetches API keys from API Gateway (if the cache is empty) and finds the key which name is
// same as qnPlanSlug
func FindBySlug(apiGatewayClient *apigateway.Client, qnPlanSlug string) (keyValue string, err error) {
	var keys []types.ApiKey
	if len(planSlugToApiKey) == 0 {
		keys, err = fetchApiKeys(apiGatewayClient)
		if err != nil {
			return
		}
		cacheApiKeys(keys)
	} else {
		log.Println("plans are already fetched")
	}
	return findPlanApiKey(qnPlanSlug, planSlugToApiKey)
}

// fetchApiKeys fetches API keys from API Gateway
func fetchApiKeys(apiGatewayClient *apigateway.Client) (keys []types.ApiKey, err error) {
	keysOutput, err := apiGatewayClient.GetApiKeys(context.TODO(), &apigateway.GetApiKeysInput{
		IncludeValues: aws.Bool(true),
	})
	if err != nil {
		err = fmt.Errorf("cannot get api keys: %w", err)
		return
	}
	keys = keysOutput.Items
	return
}

// cacheApiKeys extracts key values and saves them to the cache map
func cacheApiKeys(keys []types.ApiKey) {
	log.Println("loading plans into cache")

	for _, apiKey := range keys {
		if !apiKey.Enabled {
			log.Println("findPlanApiKey: ommiting disabled key:", apiKey.Name, apiKey.Id)
			continue
		}
		planSlugToApiKey[*apiKey.Name] = *apiKey.Value
	}
	return
}

// findPlanApiKey performs key lookup, caching the keys if the cache is empty
func findPlanApiKey(qnPlanSlug string, planToKey map[string]string) (keyValue string, err error) {
	keyValue = planToKey[qnPlanSlug]
	if keyValue == "" {
		return "", fmt.Errorf("cannot find API key for qn plan slug '%s'", qnPlanSlug)
	}
	return
}
