package config

var defaultConfig = map[string]any{
	"QnProvision.TableName": "trueblocks-qn-users-qn",
	"Chains.Allowed": map[string][]string{
		"ethereum": {"mainnet"},
	},
}
