local:
# Run the whole cloud stack locally. In order for `scanner`
# lambda to work, there must be blooms and index chunks uploaded
# to S3 bucket
	sam local start-api

test-unit:
	find . -name go.mod -execdir go test ./... \;

test:
# runs integration tests
# docker pull command is here to ensure that PostgreSQL image is downloaded before we run the tests
	docker pull --quiet postgres:15.4
	sam build --template deployment/sam/key.yml
	go test -timeout 5m -tags integration ./test/integration

build:
	sam build --config-env production --template deployment/sam/key.yml

DEPLOY_ENV = staging

deploy-staging: export DEPLOY_ENV = staging
deploy-staging: deploy

deploy-production: export DEPLOY_ENV = production
deploy-production: deploy

deploy:
# deploys the whole stack to AWS. It uses settings saved in `samconfig.toml`.
# If that file is missing, you have to call `sam deploy --guided` first.
# https://github.com/aws/aws-sam-cli/issues/4653
	@echo "Building and packaging KΞY stack template..."
	sam build --config-env $(DEPLOY_ENV) --template deployment/sam/key.yml
	sam package --config-env $(DEPLOY_ENV) --profile key-stackset-deployer --region us-east-1 --resolve-s3 --output-template-file deployment/sam/packaged_key_template.yml

# Because StackSets doesn't currently (Dec 2023) fully work with CloudFormation `Template:` macros (which are in turn required
# by SAM), we will first upload the template to CF in order to translate it from SAM to CloudFormation (Template macros will be
# resolved and we will get macro-free template back). Note `--no-execute-changeset` parameter, which blocks sam deploy from
# creating the resources (we just want to translate the template, nothing more). We also have to store the translated template's
# ARN in a temp file
	@echo "Translating SAM template to CloudFormation..."
	sam deploy --template deployment/sam/packaged_key_template.yml  --config-env $(DEPLOY_ENV) --profile key-stackset-deployer --region us-east-1 --no-execute-changeset --stack-name key-test --resolve-s3 --capabilities CAPABILITY_IAM CAPABILITY_NAMED_IAM \
	| tee \
	| grep --only-matching 'arn:aws:cloudformation:[a-z0-9\-]\+:\d\+:changeSet.*' - \
	> /tmp/key_cf_arn.txt

# We can now download the translated template using its ARN from the above
	aws --profile key-stackset-deployer cloudformation get-template --query TemplateBody --change-set-name `cat /tmp/key_cf_arn.txt` \
	> deployment/cloudformation/template.json

# We can build stackset.yml normally, because it doesn't have any `Template:` macros. It references the translated template file
# downloaded in the previous step
	@echo "Building and deploying StackSet"
	sam build --config-env $(DEPLOY_ENV) --template deployment/sam/stackset.yml
	sam deploy --config-env $(DEPLOY_ENV) --profile key-stackset-deployer --region us-east-1

set-stack-policy:
# Sets stack policy to one that prevents database removal and replacement
# The stack has to be deployed and you need to set correct user with
# AWS_PROFILE= env variable when calling this job
	$(info Using user $(AWS_PROFILE))
	$(info If you run into errors, make sure the above user is correct)
	export AWS_PROFILE
	go build -o bin/stack_policy ./tools/stack_policy
	@echo Setting policy
	bin/stack_policy --stack-name StackSet-key-prod-f6024a86-e5dc-4b96-b88e-6b6890b53f04

# Deploy Direct Customers frontend
DC_ENV = staging
DC_USER = key-staging-sso
dc-frontend-staging: export DC_ENV = staging
dc-frontent-staging: export DC_USER = key-staging-sso
dc-frontend-staging: dc-frontend

dc-frontend-production: export DC_ENV = production
dc-frontent-production: export DC_USER = key-production-sso
dc-frontend-production: dc-frontend

dc-frontend:
	echo "export const ENV='${DC_ENV}';" > /tmp/env.js
	aws --profile ${DC_USER} s3 cp --recursive direct_customers/frontend s3://key-${DC_ENV}-dc-frontend-2/frontend
	aws --profile ${DC_USER} s3 cp /tmp/env.js s3://key-${DC_ENV}-dc-frontend-2/frontend
