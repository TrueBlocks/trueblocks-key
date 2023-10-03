package db

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Connection struct {
	Host      string
	Port      int
	User      string
	Password  string
	Database  string
	db        *gorm.DB
	batchSize int
}

func (c *Connection) Connect() (err error) {
	if err := c.validate(); err != nil {
		return err
	}
	if c.batchSize == 0 {
		c.batchSize = 5000
	}
	c.db, err = gorm.Open(postgres.Open(c.dsn()), &gorm.Config{
		CreateBatchSize: c.batchSize,
		Logger:          logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		return fmt.Errorf("connection.Connect: %w", err)
	}
	return
}

func (c *Connection) Db() *gorm.DB {
	return c.db
}

func (c *Connection) String() string {
	var pass string
	if len(c.Password) > 0 {
		pass = "xxx"
	}
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", c.Host, c.User, pass, c.Database, c.Port)
}

func (c *Connection) AutoMigrate() error {
	return c.db.AutoMigrate(
		&Appearance{},
		&Progress{},
	)
}

func (c *Connection) BatchSize() int {
	return c.batchSize
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

	return nil
}
