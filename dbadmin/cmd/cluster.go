/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/spf13/cobra"
)

// clusterCmd represents the cluster command
var clusterCmd = &cobra.Command{
	Use:   "cluster",
	Short: "Cluster appearances table using appearances order index",
	RunE: func(cmd *cobra.Command, args []string) error {
		if a := YesNoPrompt(fmt.Sprintf("Cluster appearances for chain %s? WARN: it means database DOWNTIME\n", dbConn.Chain)); !a {
			log.Println("exit")
			return nil
		}

		log.Println("clustering appearances (it takes time)")
		stmt := fmt.Sprintf(
			`CLUSTER VERBOSE %s USING %s`,
			pgx.Identifier.Sanitize(pgx.Identifier{dbConn.AppearancesTableName()}),
			pgx.Identifier.Sanitize(pgx.Identifier{dbConn.AppearancesTableName() + "_appearances_order"}),
		)
		log.Println(stmt)
		if _, err := dbConn.Db().Exec(context.TODO(), stmt); err != nil {
			return err
		}

		log.Println("done")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(clusterCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clusterCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clusterCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
