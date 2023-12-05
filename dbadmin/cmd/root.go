/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	database "github.com/TrueBlocks/trueblocks-key/database/pkg"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "dbadmin",
	Short: "Database administration tool to perform one-off tasks",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if dbConn.Chain == "" {
			return errors.New("chain required")
		}

		log.Println(dbConn.String())
		if err := dbConn.Connect(context.TODO()); err != nil {
			return err
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

var dbConn *database.Connection

func init() {
	dbConn = &database.Connection{}
	rootCmd.PersistentFlags().StringVarP(&dbConn.Host, "host", "H", "localhost", "PostgreSQL host")
	rootCmd.PersistentFlags().IntVarP(&dbConn.Port, "port", "p", 5432, "PostgreSQL port")
	rootCmd.PersistentFlags().StringVarP(&dbConn.User, "user", "U", "postgres", "PostgreSQL user")
	rootCmd.PersistentFlags().StringVarP(&dbConn.Password, "password", "w", "", "PostgreSQL password")
	rootCmd.PersistentFlags().StringVarP(&dbConn.Database, "database", "d", "index", "PostgreSQL database name")
	rootCmd.PersistentFlags().StringVarP(&dbConn.Chain, "chain", "c", "", "chain")
}

func YesNoPrompt(question string) bool {
	choices := "y/N"

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", question, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		return false
	}
}
