/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"trueblocks.io/uploader/internal/db"
	"trueblocks.io/uploader/internal/export"
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

	conn, err := db.Connection(configPath)
	if err != nil {
		return err
	}

	log.Println(conn)

	return export.Export(conn, args[0])
}
