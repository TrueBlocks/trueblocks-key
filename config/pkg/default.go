package config

var defaultConfig = map[string]any{
	"QnProvision.TableName": "key-users-qn",
	"Chains.Allowed": map[string][]string{
		"ethereum": {"mainnet"},
	},
	"Convert.BatchSize":      100,
	"Convert.MaxConnections": 20,
}
