/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create database tables with `chain_` prefix",
	RunE: func(cmd *cobra.Command, args []string) error {
		if a := YesNoPrompt(fmt.Sprintf("Create tables for chain %s?\n", dbConn.Chain)); !a {
			log.Println("exit")
			return nil
		}

		log.Println("creating tables...")
		if err := dbConn.Setup(); err != nil {
			return err
		}

		log.Println("done")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}
