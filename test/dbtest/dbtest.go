package dbtest

import (
	"context"
	"fmt"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var containerName = "test-db"
var containerPort = "5432"
var containerDbUser = "postgres"
var containerDbPassword = "example"
var containerDbName = "testdb"
var containerNetwork = "testdb-network"

func ContainerNetwork() string {
	return containerNetwork
}

func ConnectionEnvs() map[string]string {
	return map[string]string{
		"KY_DATABASE_DEFAULT_HOST":     containerName,
		"KY_DATABASE_DEFAULT_PORT":     containerPort,
		"KY_DATABASE_DEFAULT_USER":     containerDbUser,
		"KY_DATABASE_DEFAULT_PASSWORD": containerDbPassword,
		"KY_DATABASE_DEFAULT_DATABASE": containerDbName,
	}
}

type TestLogConsumer struct{}

func (g *TestLogConsumer) Accept(l testcontainers.Log) {
	fmt.Println("== container log ==", string(l.Content))
}

func NewTestConnection() (conn *database.Connection, done func() error, err error) {
	dockerNetwork := ContainerNetwork()
	conn = &database.Connection{
		Database: containerDbName,
		User:     containerDbUser,
		Password: containerDbPassword,
		Chain:    "mainnet",
	}
	ctx := context.Background()
	dbContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				// Make sure image version matches one in Makefile (for target "test")
				Image:        "postgres:15.4",
				Name:         containerName,
				ExposedPorts: []string{containerPort},
				WaitingFor:   wait.ForListeningPort(nat.Port(containerPort)),
				Env: map[string]string{
					"POSTGRES_DB":       conn.Database,
					"POSTGRES_USER":     conn.User,
					"POSTGRES_PASSWORD": conn.Password,
				},
				Networks: []string{dockerNetwork},
				// For logging only:
				// Cmd: []string{"postgres", "-c", "log_statement=all", "-c", "log_destination=stderr"},
			},
			Started: true,
		},
	)
	if err != nil {
		return
	}
	terminateContainer := func() error {
		return dbContainer.Terminate(ctx)
	}

	// lc := &TestLogConsumer{}
	// dbContainer.FollowOutput(lc)
	// _ = dbContainer.StartLogProducer(ctx)

	port, err := dbContainer.MappedPort(context.Background(), nat.Port(containerPort))
	if err != nil {
		terminateContainer()
		return
	}
	conn.Port = port.Int()
	conn.Host, err = dbContainer.Host(context.Background())
	if err != nil {
		terminateContainer()
		return
	}
	if err = conn.Connect(ctx); err != nil {
		terminateContainer()
		return
	}

	err = conn.Setup()
	done = func() error {
		defer conn.Close(context.TODO())
		return terminateContainer()
	}
	return
}
