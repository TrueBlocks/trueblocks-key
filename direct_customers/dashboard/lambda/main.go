package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	keyConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
)

var ErrBadRequest = errors.New("bad request")
var ErrServerError = errors.New("internal server error")

var cnf *keyConfig.ConfigFile
var client *cognito.Client
var dynamoClient *dynamodb.Client

var jwksUrl string
var jwksCache *jwk.Cache

func init() {
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalln("error reading config:", err)

	}

	client = cognito.NewFromConfig(awsConfig)
	dynamoClient = dynamodb.NewFromConfig(awsConfig)

	cnf, err = keyConfig.Get("")
	if err != nil {
		log.Fatalln("loading config:", err)
	}

	if cnf.DirectCustomers.PoolId == "" {
		log.Fatalln("pool id needs to be provided")
	}

	jwksCache = jwk.NewCache(context.TODO())
	jwksUrl = fmt.Sprintf(
		"https://cognito-idp.us-east-1.amazonaws.com/%s/.well-known/jwks.json",
		cnf.DirectCustomers.PoolId,
	)
	if err := jwksCache.Register(jwksUrl); err != nil {
		log.Fatalln("registering jwks cache:", err)
	}
	if _, err := jwksCache.Refresh(context.TODO(), jwksUrl); err != nil {
		log.Fatalln("refreshing jwks:", err)
	}
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	log.Println("serving", request.Path)

	return handleAuth(ctx, request)
}

type TokenEndpointResult struct {
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthResponse struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Endpoint string `json:"endpoint"`
}

func handleAuth(ctx context.Context, request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, err error) {
	code := request.QueryStringParameters["code"]
	if code == "" {
		log.Println("code missing")
		err = ErrBadRequest
		return
	}

	payload := url.Values{}
	payload.Set("grant_type", "authorization_code")
	payload.Set("code", code)
	payload.Set("redirect_uri", cnf.DirectCustomers.ApiCallbackUrl)

	tokenEndpointUrl, err := url.JoinPath("https://"+cnf.DirectCustomers.PoolDomain+".auth.us-east-1.amazoncognito.com", "/oauth2/token")
	if err != nil {
		log.Println("url join path:", err)
		err = ErrServerError
		return
	}

	req, err := http.NewRequest(http.MethodPost, tokenEndpointUrl, strings.NewReader(payload.Encode()))
	if err != nil {
		log.Println("new request:", err)
		err = ErrServerError
		return
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(cnf.DirectCustomers.PoolClientId, cnf.DirectCustomers.PoolClientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("token request:", err)
		err = ErrServerError
		return
	}
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("reading response body:", err)
		err = ErrServerError
		return
	}
	resp.Body.Close()

	tokenEndpoint := &TokenEndpointResult{}
	if err = json.Unmarshal(responseBody, tokenEndpoint); err != nil {
		log.Println("unmarshal response body:", err)
		err = ErrServerError
		return
	}
	log.Println("ID Token:", tokenEndpoint.IdToken)
	log.Println("Response:", string(responseBody))

	// Verify the token
	keyset, err := jwksCache.Get(ctx, jwksUrl)
	if err != nil {
		log.Println("getting jwks:", err)
		err = ErrServerError
		return
	}
	token, err := jwt.Parse([]byte(tokenEndpoint.IdToken), jwt.WithKeySet(keyset))
	if err != nil {
		log.Println("parsing JWT token:", err)
		err = ErrServerError
		return
	}
	// Now get user details
	email, ok := token.Get("email")
	if !ok {
		log.Println("no email")
		err = ErrServerError
		return
	}

	username, ok := token.Get("cognito:username")
	if !ok {
		log.Println("no username")
		err = ErrServerError
		return
	}

	// find endpoint
	endpoint, err := findEndpoint(ctx, email.(string))
	if err != nil {
		log.Println("finding endpoint:", err)
		err = ErrServerError
		return
	}

	authResponse := AuthResponse{
		Email:    email.(string),
		Username: username.(string),
		Endpoint: endpoint,
	}
	encodedResponse, err := json.Marshal(&authResponse)
	if err != nil {
		log.Println("encoding response:", err)
		err = ErrServerError
		return
	}
	log.Println("Responding with", string(encodedResponse))
	response.Body = string(encodedResponse)
	response.StatusCode = 200
	response.Headers = map[string]string{
		"Access-Control-Allow-Headers": "*",
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
	}
	return
}

func findEndpoint(ctx context.Context, email string) (endpoint string, err error) {
	output, err := dynamoClient.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(cnf.DirectCustomers.TableName),
		IndexName:              aws.String(cnf.DirectCustomers.EmailIndexName),
		KeyConditionExpression: aws.String("Email = :email"),
		ExpressionAttributeValues: map[string]dynamodbTypes.AttributeValue{
			":email": &dynamodbTypes.AttributeValueMemberS{
				Value: email,
			},
		},
	})
	if err != nil {
		return
	}
	if l := len(output.Items); l != 1 {
		err = fmt.Errorf("expected 1 endpoint, but found %d", l)
		return
	}

	err = attributevalue.Unmarshal(output.Items[0]["EndpointId"], &endpoint)
	return
}

func main() {
	lambda.Start(HandleRequest)
}
