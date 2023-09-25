local:
# Run the whole cloud stack locally. In order for `scanner`
# lambda to work, there must be blooms and index chunks uploaded
# to S3 bucket
	sam local start-api

deploy:
# deploys the whole stack to AWS. It uses settings saved in `samconfig.toml`.
# If that file is missing, you have to call `sam deploy --guided` first.
	sam build
	sam deploy

deploy-dev:
	sam build
	sam deploy --on-failure DELETE
