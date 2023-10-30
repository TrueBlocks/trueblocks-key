package dbtest

import (
	"context"

	"github.com/docker/go-connections/nat"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	database "trueblocks.io/database/pkg"
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
		"QNEXT_DATABASE_DEFAULT_HOST":     containerName,
		"QNEXT_DATABASE_DEFAULT_PORT":     containerPort,
		"QNEXT_DATABASE_DEFAULT_USER":     containerDbUser,
		"QNEXT_DATABASE_DEFAULT_PASSWORD": containerDbPassword,
		"QNEXT_DATABASE_DEFAULT_DATABASE": containerDbName,
	}
}

func NewTestConnection() (conn *database.Connection, done func() error, err error) {
	dockerNetwork := ContainerNetwork()
	conn = &database.Connection{
		Database: containerDbName,
		User:     containerDbUser,
		Password: containerDbPassword,
	}
	ctx := context.Background()
	dbContainer, err := testcontainers.GenericContainer(
		ctx,
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "postgres:latest",
				Name:         containerName,
				ExposedPorts: []string{containerPort},
				WaitingFor:   wait.ForListeningPort(nat.Port(containerPort)),
				Env: map[string]string{
					"POSTGRES_DB":       conn.Database,
					"POSTGRES_USER":     conn.User,
					"POSTGRES_PASSWORD": conn.Password,
				},
				Networks: []string{dockerNetwork},
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
	if err = conn.Connect(); err != nil {
		terminateContainer()
		return
	}

	err = conn.AutoMigrate()
	done = terminateContainer
	return
}
