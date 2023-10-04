/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"trueblocks.io/uploader/internal/convert"
	"trueblocks.io/uploader/internal/db"
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

	conn, err := db.Connection(configPath, dbConfigKey)
	if err != nil {
		return err
	}

	log.Println(conn)
	return convert.Convert(conn, args[0])
}
