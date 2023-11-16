/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/TrueBlocks/trueblocks-key/extract/internal/db"
	"github.com/TrueBlocks/trueblocks-key/extract/internal/export"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export destination/path",
	Short: "Exports local database table to CSV files that can be uploaded to servers",
	Args:  cobra.ExactArgs(1),
	RunE:  runExport,
}

func init() {
	rootCmd.AddCommand(exportCmd)
}

func runExport(cmd *cobra.Command, args []string) error {
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

	return export.Export(conn, args[0])
}
