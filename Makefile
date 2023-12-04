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
	@echo "Building and packaging KΞY stack template..."
	sam build --config-env production --template deployment/sam/key.yml
	sam package --config-env production --profile key-stackset-deployer --region us-east-1 --resolve-s3 --output-template-file deployment/sam/packaged_key_template.yml

# Because StackSets doesn't currently (Dec 2023) fully work with CloudFormation `Template:` macros (which are in turn required
# by SAM), we will first upload the template to CF in order to translate it from SAM to CloudFormation (Template macros will be
# resolved and we will get macro-free template back). Note `--no-execute-changeset` parameter, which blocks sam deploy from
# creating the resources (we just want to translate the template, nothing more). We also have to store the translated template's
# ARN in a temp file
	@echo "Translating SAM template to CloudFormation..."
	sam deploy --template deployment/sam/packaged_key_template.yml  --config-env production --profile key-stackset-deployer --region us-east-1 --no-execute-changeset --stack-name key-test --resolve-s3 --capabilities CAPABILITY_IAM \
	| tee \
	| grep --only-matching 'arn:aws:cloudformation:[a-z0-9\-]\+:\d\+:changeSet.*' - \
	> /tmp/key_cf_arn.txt

# We can now download the translated template using its ARN from the above
	aws --profile key-stackset-deployer cloudformation get-template --query TemplateBody --change-set-name `cat /tmp/key_cf_arn.txt` \
	> deployment/cloudformation/template.json

# We can build stackset.yml normally, because it doesn't have any `Template:` macros. It references the translated template file
# downloaded in the previous step
	@echo "Building and deploying StackSet"
	sam build --config-env production --template deployment/sam/stackset.yml
	sam deploy --config-env production --profile key-stackset-deployer --region us-east-1
