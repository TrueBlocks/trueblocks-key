/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"log"

	config "github.com/TrueBlocks/trueblocks-key/config/pkg"
	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/TrueBlocks/trueblocks-key/extract/internal/convert"
	"github.com/spf13/cobra"
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert path/to/index",
	Short: "Converts Unchained Index chunks to SQL",
	Args:  cobra.ExactArgs(1),
	RunE:  runConvert,
}

func init() {
	convertCmd.Flags().StringP("insert_url", "", "", "URL of insert tool")
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

	insertServerUrl, err := cmd.Flags().GetString("insert_url")
	if err != nil {
		return err
	}

	cnf, err := config.Get(configPath)
	if err != nil {
		return err
	}

	var receiver convert.AppearanceReceiver

	if insertServerUrl == "" {
		dbRecv := &convert.DatabaseReceiver{
			DbConn: &database.Connection{
				Host:     cnf.Database[dbConfigKey].Host,
				Port:     cnf.Database[dbConfigKey].Port,
				Database: cnf.Database[dbConfigKey].Database,
				User:     cnf.Database[dbConfigKey].User,
				Password: cnf.Database[dbConfigKey].Password,
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
