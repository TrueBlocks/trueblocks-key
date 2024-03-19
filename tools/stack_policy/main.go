package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

var client *cloudformation.Client

func fatalln(v ...any) {
	log.Println(v...)
	log.Fatalln("NOTE: make sure to set the correct user using AWS_PROFILE env variable")
}

func handleListStacks() {
	result, err := client.ListStacks(context.TODO(), &cloudformation.ListStacksInput{
		StackStatusFilter: []types.StackStatus{types.StackStatusCreateComplete, types.StackStatusUpdateComplete},
	})
	if err != nil {
		fatalln(err)
	}
	for _, summary := range result.StackSummaries {
		fmt.Println(*summary.StackName)
	}
}

func buildPolicy(dbLogicalId string) string {
	return fmt.Sprintf(`
{
  "Statement" : [
	{
      "Effect" : "Allow",
      "Action" : "Update:*",
      "Principal": "*",
      "Resource" : "*"
    },
    {
      "Effect" : "Deny",
      "Action" : ["Update:Replace", "Update:Delete"],
      "Principal": "*",
      "Resource" : "LogicalResourceId/%[1]s"
    }
  ]
}`,
		dbLogicalId,
	)
}

func handleGetPolicy(stackName string) {
	result, err := client.GetStackPolicy(
		context.TODO(),
		&cloudformation.GetStackPolicyInput{
			StackName: &stackName,
		},
	)
	if err != nil {
		fatalln(err)
	}

	fmt.Println(*result.StackPolicyBody)
}

func handleSetPolicy(stackName string, dbLogicalId string) {
	policy := buildPolicy(dbLogicalId)
	_, err := client.SetStackPolicy(
		context.TODO(),
		&cloudformation.SetStackPolicyInput{
			StackName:       &stackName,
			StackPolicyBody: &policy,
		},
	)
	if err != nil {
		fatalln(err)
	}
}

func main() {
	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	client = cloudformation.NewFromConfig(cfg)

	var list bool
	var get bool
	var stackName string
	var dbLogicalId string
	flag.BoolVar(&list, "list", false, "only list stacks, do not set policy")
	flag.BoolVar(&get, "get", false, "instead of settings a policy, retrieve it")
	flag.StringVar(&stackName, "stack-name", "", "the name of the stack to update (required)")
	flag.StringVar(&dbLogicalId, "db", "IndexDatabase", "database LogicalId (name from SAM template). Default: IndexDatabase")
	flag.Parse()

	if list {
		handleListStacks()
		return
	}

	if stackName == "" {
		fatalln("--stack-name required. Use --list to list the stacks and AWS_PROFILE to set the user")
	}

	if !get {
		handleSetPolicy(stackName, dbLogicalId)
	} else {
		handleGetPolicy(stackName)
	}
}
