package main

import (
	"context"
	"flag"
	"log"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	config "github.com/TrueBlocks/trueblocks-key/config/pkg"
	"github.com/TrueBlocks/trueblocks-key/queue/insert/internal/queue"
	"github.com/TrueBlocks/trueblocks-key/queue/insert/internal/server"
)

var configPath string
var port int
var file string

var client *sqs.Client

func main() {
	flag.StringVar(&configPath, "config", "", "path to configuration file")
	flag.StringVar(&file, "file", "", "(testing only) use this local file instead of a real queue")
	flag.IntVar(&port, "port", 5555, "port to listen on")
	flag.Parse()

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	client := sqs.NewFromConfig(cfg)

	keyConfig, err := config.Get(configPath)
	if err != nil {
		log.Fatalln("reading configuration:", err)
	}

	var impl queue.RemoteQueuer
	if file != "" {
		impl = queue.NewFileQueue(file)
	} else {
		if keyConfig.Sqs.QueueName == "" {
			log.Fatalln("Cannot read QueueName. Either use --config or set env variable")
		}
		log.Println("Loading SQS queue", keyConfig.Sqs.QueueName)
		impl = queue.NewSqsQueue(client, keyConfig)
	}

	q, err := queue.NewQueue(impl)
	if err != nil {
		log.Fatalln(err)
	}
	srv := server.New(q)

	if err := srv.Start(port); err != nil {
		log.Fatalln(err)
	}
}
