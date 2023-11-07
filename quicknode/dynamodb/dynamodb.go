package qnDynamodb

import "os"

var LocalDynamoDbSettings = struct {
	ContainerName string
	Port          string
}{
	"dynamodb",
	"8000",
}

func ShouldUseLocal() bool {
	return os.Getenv("AWS_SAM_LOCAL") == "true"
}
