local:
# Run the whole cloud stack locally. In order for `scanner`
# lambda to work, there must be blooms and index chunks uploaded
# to S3 bucket
	sam local start-api

test:
# runs integration tests
# docker pull command is here to ensure that PostgreSQL image is downloaded before we run the tests
	docker pull --quiet postgres:15.4
	sam build --template deployment/sam/key.yml
	go test -timeout 5m -tags integration ./test/integration

deploy:
# deploys the whole stack to AWS. It uses settings saved in `samconfig.toml`.
# If that file is missing, you have to call `sam deploy --guided` first.
# https://github.com/aws/aws-sam-cli/issues/4653
	sam build --config-env production --template deployment/sam/key.yml
	sam package --config-env production --profile samqndeployer --region us-east-1 --resolve-s3 --output-template-file deployment/sam/packaged_key_template.yml
	sam build --config-env production --template deployment/sam/stackset.yml
	sam deploy --config-env production --profile samqndeployer --region us-east-1

deploy-stack:
	sam build --config-env production --template deployment/sam/stackset.yml
	sam deploy --config-env production --profile samqndeployer --region us-east-1

# deploy-dev:
# 	sam build
# 	sam deploy --profile samqndeployer --region us-east-1 --on-failure DELETE
