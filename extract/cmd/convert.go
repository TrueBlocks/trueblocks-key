/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"log"

	"github.com/spf13/cobra"
	config "trueblocks.io/config/pkg"
	database "trueblocks.io/database/pkg"
	"trueblocks.io/uploader/internal/convert"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert path/to/index",
	Short: "Converts Unchained Index chunks to SQL",
	Args:  cobra.ExactArgs(1),
	RunE:  runConvert,
}

func init() {
	rootCmd.AddCommand(convertCmd)
}

func runConvert(cmd *cobra.Command, args []string) error {
	configPath, err := cmd.Flags().GetString("config_path")
	if err != nil {
		return err
	}

	dbConfigKey, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}

	cnf, err := config.Get(configPath)
	if err != nil {
		return err
	}

	var receiver convert.AppearanceReceiver

	if dbConfigKey != "" {
		dbRecv := &convert.DatabaseReceiver{
			DbConn: &database.Connection{
				Host:     cnf.Database["default"].Host,
				Port:     cnf.Database["default"].Port,
				Database: cnf.Database["default"].Database,
				User:     cnf.Database["default"].User,
				Password: cnf.Database["default"].Password,
			},
		}
		log.Println(dbRecv.DbConn)
		if err := dbRecv.DbConn.Connect(); err != nil {
			return err
		}
		if err := dbRecv.DbConn.AutoMigrate(); err != nil {
			return err
		}
		receiver = dbRecv
	} else {
		insertServerUrl, err := cmd.Flags().GetString("insert_url")
		if err != nil {
			return err
		}
		if insertServerUrl == "" {
			return errors.New("--insert_url or --database required")
		}
		receiver = &convert.QueueReceiver{
			InsertUrl:      insertServerUrl,
			MaxConnections: cnf.Convert.MaxConnections,
		}
	}

	return convert.Convert(cnf, receiver, args[0])
}
