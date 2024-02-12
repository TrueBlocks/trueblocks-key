package dbtest

import (
	"context"
	"fmt"
	"time"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
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
	dbContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15.4"),
		postgres.WithDatabase(conn.Database),
		postgres.WithUsername(conn.User),
		postgres.WithPassword(conn.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(1*time.Minute)),
		testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Name:     "test-db",
				Networks: []string{dockerNetwork},
			},
		}),
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
