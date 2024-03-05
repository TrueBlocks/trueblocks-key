package database

import (
	"context"
	"fmt"

	"github.com/TrueBlocks/trueblocks-key/database/pkg/sql"
	"github.com/jackc/pgx/v5"
)

type Connection struct {
	Host      string
	Port      int
	User      string
	Password  string
	Database  string
	Chain     string
	conn      *pgx.Conn
	batchSize int
}

func (c *Connection) Connect(ctx context.Context) (err error) {
	if err := c.validate(); err != nil {
		return err
	}
	if c.batchSize == 0 {
		c.batchSize = 5000
	}
	connConfig, err := pgx.ParseConfig(c.dsn())
	if err != nil {
		return fmt.Errorf("connection.Connect: parsing db config: %w", err)
	}
	// TODO: enabling this can protect us from RDS Proxy pinning, but it has downsides.
	// TODO: Is it needed?
	connConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	// c.conn, err = pgx.Connect(ctx, c.dsn())
	c.conn, err = pgx.ConnectConfig(ctx, connConfig)
	if err != nil {
		return fmt.Errorf("connection.Connect: %w", err)
	}
	return
}

func (c *Connection) Close(ctx context.Context) error {
	return c.conn.Close(ctx)
}

func (c *Connection) Db() *pgx.Conn {
	return c.conn
}

func (c *Connection) String() string {
	var pass string
	if len(c.Password) > 0 {
		pass = "xxx"
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.Host, c.User, pass, c.Database, c.Port)
}

func (c *Connection) BatchSize() int {
	return c.batchSize
}

func (c *Connection) Setup() (err error) {
	if _, err = c.conn.Exec(context.TODO(), sql.CreateTableAddresses(c.AddressesTableName())); err != nil {
		return fmt.Errorf("creating address table (%s): %w", c.Chain, err)
	}
	if _, err = c.conn.Exec(context.TODO(), sql.CreateTableAppearances(c.AppearancesTableName(), c.AddressesTableName())); err != nil {
		return fmt.Errorf("creating appearances table (%s): %w", c.Chain, err)
	}
	if _, err = c.conn.Exec(context.TODO(), sql.CreateAppearancesOrderIndex(c.AppearancesTableName())); err != nil {
		return fmt.Errorf("creating appearances order index (%s): %w", c.Chain, err)
	}

	if _, err = c.conn.Exec(context.TODO(), sql.CreateTableChunks(c.ChunksTableName())); err != nil {
		return fmt.Errorf("creating chunks table (%s): %w", c.Chain, err)
	}
	return nil
}

func (c *Connection) CountAppearances() (count int, err error) {
	rows, err := c.conn.Query(
		context.TODO(),
		fmt.Sprintf(
			"select count(*) from %s",
			pgx.Identifier.Sanitize(pgx.Identifier{c.AppearancesTableName()}),
		),
	)
	if err != nil {
		return
	}
	count, err = pgx.CollectOneRow[int](rows, pgx.RowTo[int])
	return
}

func (c *Connection) dsn() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.Host, c.User, c.Password, c.Database, c.Port)
}

func (c *Connection) validate() error {
	// Ports below 1024 are reserved, so it would be strange
	// if database ran there
	if c.Port < 1024 {
		return fmt.Errorf("invalid database port %d", c.Port)
	}

	// We need host, even if it's localhost
	if c.Host == "" {
		return fmt.Errorf("database host missing")
	}

	// We need database name
	if c.Database == "" {
		return fmt.Errorf("database name missing")
	}

	if c.Chain == "" {
		return fmt.Errorf("chain empty")
	}

	return nil
}
