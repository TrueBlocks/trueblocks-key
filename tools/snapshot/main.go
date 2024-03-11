package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go/aws"
)

var snapshotList bool
var snapshotCreate bool
var snapshotShare bool
var snapshotShareSnapshotName string
var snapshotInstanceName string
var snapshotShareAccountId string

var client *rds.Client

func main() {
	if snapshotShare {
		if !snapshotCreate && snapshotShareSnapshotName == "" {
			log.Fatalln("snapshot-name is required when using 'share', unless you are also using 'create'")
		}
		if snapshotShareAccountId == "" {
			log.Fatalln("instance-name required")
		}
	}

	if snapshotList {
		list()
		return
	}

	var snapshotName string
	if snapshotCreate {
		snapshotName = takeSnapshot(snapshotInstanceName)
	} else {
		snapshotName = snapshotShareSnapshotName
	}

	if snapshotShare {
		shareSnapshot(snapshotName, snapshotShareAccountId)
	}

	return

}

func init() {
	flag.BoolVar(&snapshotList, "list", false, "lists manual snapshots")
	flag.BoolVar(&snapshotCreate, "create", false, "creates a new snapshot")
	flag.BoolVar(&snapshotShare, "share", false, "shares a snapshot snapshot")
	flag.StringVar(&snapshotShareAccountId, "account-id", "", "account to share the snapshot with")
	flag.StringVar(&snapshotShareSnapshotName, "snapshot-name", "", "name of the snapshot to share (unless using --create)")
	flag.StringVar(&snapshotInstanceName, "instance-name", "", "name of the source instance")

	flag.Parse()

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	client = rds.NewFromConfig(cfg)
}

func takeSnapshot(instanceName string) string {
	snapshotName := "shared-" + time.Now().UTC().Format("2006-01-02T15-04-05")
	output, err := client.CreateDBSnapshot(context.TODO(), &rds.CreateDBSnapshotInput{
		DBInstanceIdentifier: aws.String(instanceName),
		DBSnapshotIdentifier: aws.String(snapshotName),
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(snapshotName, ",", *output.DBSnapshot.DBSnapshotArn)
	return snapshotName
}

func shareSnapshot(snapshotName string, accountId string) {
	_, err := client.ModifyDBSnapshotAttribute(context.TODO(), &rds.ModifyDBSnapshotAttributeInput{
		AttributeName:        aws.String("restore"),
		ValuesToAdd:          []string{accountId},
		DBSnapshotIdentifier: aws.String(snapshotName),
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func list() {
	// TODO: this returnes paginated data
	output, err := client.DescribeDBSnapshots(context.TODO(), &rds.DescribeDBSnapshotsInput{SnapshotType: aws.String("manual")})
	if err != nil {
		log.Fatalln(err)
	}
	for _, snapshot := range output.DBSnapshots {
		fmt.Println(*snapshot.DBSnapshotIdentifier, ",", *snapshot.Status, ",", *snapshot.DBSnapshotArn)
	}
}
