package dbtest

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	database "trueblocks.io/database/pkg"
)

func NewTestConnection() (conn *database.Connection, err error) {
	conn = &database.Connection{
		Database: "testdb",
		User:     "postgres",
		Password: "postgres",
	}
	dbContainer, err := testcontainers.GenericContainer(
		context.Background(),
		testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				Image:        "postgres:latest",
				ExposedPorts: []string{"5432/tcp"},
				WaitingFor:   wait.ForListeningPort("5432/tcp"),
				Env: map[string]string{
					"POSTGRES_DB":       conn.Database,
					"POSTGRES_USER":     conn.User,
					"POSTGRES_PASSWORD": conn.Password,
				},
			},
			Started: true,
		},
	)
	if err != nil {
		return
	}

	port, err := dbContainer.MappedPort(context.Background(), "5432")
	if err != nil {
		return
	}
	conn.Port = port.Int()
	conn.Host, err = dbContainer.Host(context.Background())
	if err != nil {
		return
	}
	if err = conn.Connect(); err != nil {
		return
	}

	err = conn.AutoMigrate()
	return
}
