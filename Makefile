local:
# Run the whole cloud stack locally. In order for `scanner`
# lambda to work, there must be blooms and index chunks uploaded
# to S3 bucket
	sam local start-api

test:
# runs integration tests
# docker pull command is here to ensure that PostgreSQL image is downloaded before we run the tests
	docker pull --quiet postgres:15.4
	sam build
	go test -timeout 30s -tags integration -run ^TestLambdaQnAuthorizer$ trueblocks.io/test/integration

deploy:
# deploys the whole stack to AWS. It uses settings saved in `samconfig.toml`.
# If that file is missing, you have to call `sam deploy --guided` first.
	sam build
	sam deploy --profile samqndeployer --region us-east-1

deploy-dev:
	sam build
	sam deploy --profile samqndeployer --region us-east-1 --on-failure DELETE
