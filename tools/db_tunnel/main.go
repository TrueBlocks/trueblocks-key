package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmTypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

var ssmClient *ssm.Client
var ec2Client *ec2.Client
var rdsClient *rds.Client
var smClient *secretsmanager.Client

var errExpectedSubcommand = errors.New(
	"expected subcommand: private-key, bastion-domain, rds-domain, rds-username, rds-password",
)

func init() {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	ssmClient = ssm.NewFromConfig(cfg)
	rdsClient = rds.NewFromConfig(cfg)
	ec2Client = ec2.NewFromConfig(cfg)
	smClient = secretsmanager.NewFromConfig(cfg)

	if len(os.Args) < 2 {
		log.Fatalln(errExpectedSubcommand)
	}
}

func main() {
	switch os.Args[1] {
	case "private-key":
		getPrivateKey()
	case "bastion-domain":
		getBastionDomain()
	case "rds-domain":
		getRdsEndpoint()
	case "rds-username":
		getRdsAdminUsername()
	case "rds-password":
		getRdsAdminPassword()
	default:
		log.Fatalln(errExpectedSubcommand)
	}
}

func getPrivateKey() {
	describeOutput, err := ssmClient.DescribeParameters(context.TODO(), &ssm.DescribeParametersInput{
		ParameterFilters: []ssmTypes.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Option: aws.String("Contains"),
				Values: []string{"/ec2/keypair"},
			},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}
	if len(describeOutput.Parameters) == 0 {
		log.Fatalln("found no keys")
	}
	if len(describeOutput.Parameters) > 1 {
		names := make([]string, 0, len(describeOutput.Parameters))
		for _, param := range describeOutput.Parameters {
			names = append(names, *param.Name)
		}
		log.Fatalln("found more than 1 key:", names)
	}

	getOutput, err := ssmClient.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           describeOutput.Parameters[0].Name,
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(*getOutput.Parameter.Value)
}

func getBastionDomain() {
	describeOutput, err := ec2Client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		Filters: []ec2Types.Filter{
			{
				Name:   aws.String("instance-state-name"),
				Values: []string{"running"},
			},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	if len(describeOutput.Reservations) > 1 {
		log.Fatalln("found more than 1 reservation")
	}

	instances := describeOutput.Reservations[0].Instances
	if len(instances) == 0 {
		log.Fatalln("found no instances")
	}
	if len(instances) > 1 {
		log.Fatalln("found more than 1 instance running")
	}
	fmt.Println(*instances[0].PublicDnsName)
}

func getRdsEndpoint() {
	instance, err := findRdsInstance()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(*instance.Endpoint.Address)
}

func getRdsAdminUsername() {
	secret, err := findRdsAdminSecret()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(secret.Username)
}

func getRdsAdminPassword() {
	secret, err := findRdsAdminSecret()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(secret.Password)
}

type awsSecret struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func findRdsAdminSecret() (secret *awsSecret, err error) {
	instance, err := findRdsInstance()
	if err != nil {
		return
	}
	secretArn := instance.MasterUserSecret.SecretArn
	getOutput, err := smClient.GetSecretValue(context.TODO(), &secretsmanager.GetSecretValueInput{
		SecretId: secretArn,
	})
	if err != nil {
		return
	}
	secret = &awsSecret{}
	err = json.Unmarshal([]byte(*getOutput.SecretString), secret)
	return
}

func findRdsInstance() (instance *rdsTypes.DBInstance, err error) {
	describeOutput, err := rdsClient.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if err != nil {
		return
	}

	if len(describeOutput.DBInstances) == 0 {
		err = errors.New("found no instances")
		return
	}
	if len(describeOutput.DBInstances) > 1 {
		err = errors.New("found more than 1 database instance")
		return
	}

	instance = &describeOutput.DBInstances[0]
	return
}
