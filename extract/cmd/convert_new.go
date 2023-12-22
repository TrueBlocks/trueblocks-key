package cmd

import (
	"fmt"

	config "github.com/TrueBlocks/trueblocks-key/config/pkg"
	convertNew "github.com/TrueBlocks/trueblocks-key/extract/internal/convert_new"
	"github.com/spf13/cobra"
)

var convertNewCmd = &cobra.Command{
	Use:   "convert_new path/to/index",
	Short: "New fast tool to convert Unchained Index chunks to SQL",
	Args:  cobra.ExactArgs(1),
	RunE:  runConvertNew,
}

func init() {
	rootCmd.AddCommand(convertNewCmd)
}

func runConvertNew(cmd *cobra.Command, args []string) (err error) {
	configPath, err := cmd.Flags().GetString("config_path")
	if err != nil {
		return err
	}

	dbConfigKey, err := cmd.Flags().GetString("database")
	if err != nil {
		return err
	}
	if dbConfigKey == "" {
		dbConfigKey = "default"
	}

	cnf, err := config.Get(configPath)
	if err != nil {
		return err
	}

	host := cnf.Database[dbConfigKey].Host
	port := cnf.Database[dbConfigKey].Port
	database := cnf.Database[dbConfigKey].Database
	user := cnf.Database[dbConfigKey].User
	password := cnf.Database[dbConfigKey].Password
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, database)

	convertNew.ConvertDir(args[0], dsn)
	return nil
}
