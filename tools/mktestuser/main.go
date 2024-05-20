package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	keyConfig "github.com/TrueBlocks/trueblocks-key/config/pkg"
	"github.com/TrueBlocks/trueblocks-key/direct_customers/endpoint"
	qnaccount "github.com/TrueBlocks/trueblocks-key/quicknode/account"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	uuid "github.com/satori/go.uuid"
)

const endpointPrefix = "auto-test-"

var filePath string
var remove bool
var create bool = true
var endpointCount uint
var scenarioPath string
var keyConfigPath string

var apiKey *qnaccount.ApiKey

var dynamoClient *dynamodb.Client
var cnf *keyConfig.ConfigFile

func init() {
	flag.StringVar(&filePath, "file", "", "path to endpoints file (see below)")
	flag.BoolVar(&create, "create", true, "create endpoints and save them to the file (default)")
	flag.UintVar(&endpointCount, "count", 0, "number of endpoints to create")
	flag.BoolVar(&remove, "remove", false, "remove endpoints listed in the file")
	flag.StringVar(&scenarioPath, "scenario", "", "path of scenario file to generate using enpoint file")
	flag.StringVar(&keyConfigPath, "config", "", "path to Key configuration file")
	flag.Parse()

	if remove {
		create = false
	}

	if scenarioPath != "" {
		create = false
		if filePath == "" {
			log.Fatalln("endpoint file is required")
		}
	}

	if create {
		if endpointCount == 0 {
			log.Fatalln("wrong --count:", endpointCount)
		}
	}

	if filePath == "" {
		log.Fatalln("missing required flag: --file")
	}

	if scenarioPath != "" {
		return
	}

	var err error
	if cnf, err = keyConfig.Get(keyConfigPath); err != nil {
		log.Fatalln("cannot read Key configuration")
		return
	}

	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalln("error loading AWS configuration:", err)
	}

	dynamoClient = dynamodb.NewFromConfig(awsConfig)
	apiGatewayClient := apigateway.NewFromConfig(awsConfig)

	keysOutput, err := apiGatewayClient.GetApiKeys(context.TODO(), &apigateway.GetApiKeysInput{
		IncludeValues: aws.Bool(true),
	})
	if err != nil {
		log.Fatalln("cannot load API Gateway API keys:", err)
	}
	for _, key := range keysOutput.Items {
		if *key.Name == cnf.DirectCustomers.DefaultApiKeyName {
			apiKey = &qnaccount.ApiKey{
				Name:  *key.Name,
				Value: *key.Value,
			}
			break
		}
	}
	if apiKey == nil {
		log.Fatalln("cannot find default API Gateway API key")
	}
}

func main() {
	openFlags := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	if remove || scenarioPath != "" {
		openFlags = os.O_RDONLY
	}
	file, err := os.OpenFile(filePath, openFlags, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	if create {
		createEndpoints(file)
	}
	if remove {
		removeEndpoints(file)
	}
	if scenarioPath != "" {
		scenarioFile, err := os.OpenFile(scenarioPath, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalln("opening scenario file:", err)
		}
		defer scenarioFile.Close()

		generateTests(file, scenarioFile)
	}
}

func createEndpoints(file *os.File) {
	created := time.Now()
	for i := 0; i < int(endpointCount); i++ {
		testEndpoint := endpoint.Endpoint{
			Endpoint: endpointPrefix + uuid.NewV4().String(),
			Created:  &created,
			Email:    "auto-test@trueblocks.io",
			ApiKey:   *apiKey,
		}
		err := testEndpoint.Save(context.TODO(), dynamoClient, cnf.DirectCustomers.TableName)
		if err != nil {
			log.Fatalln("saving endpoint:", err)
		}
		if _, err := file.WriteString(testEndpoint.Endpoint + "\n"); err != nil {
			log.Fatalln("writing endpoint to the file:", err)
		}
		fmt.Println(testEndpoint.Endpoint)
	}
}

func removeEndpoints(file *os.File) {
	reader := bufio.NewReader(file)
	for {
		rawStr, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalln("reading the file:", err)
			}
		}

		endpointId := strings.TrimSpace(rawStr)
		// _, err = endpoint.Find(context.TODO(), dynamoClient, tableName, endpointId)
		// if err != nil {
		// 	log.Fatalf("cannot find endpoint '%s': %s", endpointId, err)
		// }
		encodedEndpointId, err := attributevalue.Marshal(endpointId)
		if err != nil {
			log.Fatalln("marshal endpoint id:", err)
		}
		key := map[string]types.AttributeValue{
			"EndpointId": encodedEndpointId,
		}

		_, err = dynamoClient.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
			TableName: aws.String(cnf.DirectCustomers.TableName),
			Key:       key,
		})
		if err != nil {
			log.Fatalf("deleting endpoint '%s': %s", endpointId, err)
		}
		fmt.Println(endpointId)
	}
}

func generateTests(file *os.File, testOutput *os.File) {
	bigAddressesLen := len(bigAddresses)
	currentBigAddress := 0

	scenarioHeader := fmt.Sprintf(`baseUrl = "%s"
duration = "10m"
`, cnf.Misc.DirecCustomersApiUrl)
	scenarioBody := `
[scenarios.%s]
address = "%s"
perPage = 100
directUser = "%s"
`

	if _, err := testOutput.WriteString(scenarioHeader); err != nil {
		log.Fatalln("writing scenario header:", err)
	}

	reader := bufio.NewReader(file)
	for {
		rawStr, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalln("reading the file:", err)
			}
		}

		endpointId := strings.TrimSpace(rawStr)
		address := bigAddresses[currentBigAddress]
		if currentBigAddress+1 == bigAddressesLen {
			currentBigAddress = 0
		} else {
			currentBigAddress++
		}

		_, err = testOutput.WriteString(fmt.Sprintf(
			scenarioBody,
			endpointId,
			address,
			endpointId,
		))
		if err != nil {
			log.Fatalln("writing scenario body:", err)
		}
	}

}
