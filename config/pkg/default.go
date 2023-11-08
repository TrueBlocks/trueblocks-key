package config

var defaultConfig = ConfigFile{
	QnProvision: qnProvisionGroup{
		TableName: "trueblocks-qn-users-qn",
	},
	Chains: chainsGroup{
		Allowed: map[string][]string{
			"ethereum": {"mainnet"},
		},
	},
}
