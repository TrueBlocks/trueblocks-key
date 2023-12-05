/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
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

var exportAddresses bool
var exportAppearances bool

func init() {
	exportCmd.Flags().BoolVar(&exportAddresses, "addresses", false, "export addresses table")
	exportCmd.Flags().BoolVar(&exportAppearances, "appearances", false, "export appearances table")

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

	if exportAddresses && exportAppearances {
		return errors.New("cannot export two tables at the same time")
	}

	conn, err := db.Connection(configPath, dbConfigKey)
	if err != nil {
		return err
	}

	log.Println(conn)

	if exportAddresses {
		return export.ExportAddresses(conn, args[0])
	}

	if exportAppearances {
		return export.ExportAppearances(conn, args[0])
	}

	return errors.New("specify which table to export")
}
