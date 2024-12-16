default: deploy

APP_NAME?=burrow

GIT_REVISION=$(shell git rev-parse --short HEAD)

AWS_ACCOUNT_ID=$(shell aws sts get-caller-identity --query Account --output text)

AWS_REGION?=us-east-1

LAMBDA_BUILD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

LAMBDA_BINARY=dist/burrow-$(GIT_REVISION).zip

TERRAFORM_BUCKET_NAME?=terraform-$(AWS_ACCOUNT_ID)

BURROW_BUCKET_NAME?=burrow-$(AWS_ACCOUNT_ID)

AUTO_APPROVE?=false

$(LAMBDA_BINARY): $(shell find . -name '*.go') go.mod go.sum
	mkdir -p dist
	cd cmd/lambda && $(LAMBDA_BUILD) -o ../../dist/bootstrap .
	zip -j $(LAMBDA_BINARY) dist/bootstrap

.PHONY: clean
clean:
	rm -rf dist

TF_INIT_VARS=-backend-config=bucket=$(TERRAFORM_BUCKET_NAME) \
	-backend-config=key=states/$(APP_NAME)/terraform.tfstate \
	-backend-config=region=$(AWS_REGION)

TF_VARS=-var name=$(APP_NAME) \
	-var git_revision=$(GIT_REVISION) \
	-var lambda_filename=../../$(LAMBDA_BINARY) \
	-var lambda_handler=burrow \
	-var bucket_name=$(BURROW_BUCKET_NAME)

.PHONY: deploy
deploy: $(LAMBDA_BINARY)
	cd terraform/main && \
	terraform init $(TF_INIT_VARS) -reconfigure && \
	terraform apply -auto-approve=$(AUTO_APPROVE) $(TF_VARS) && \
	terraform output -json function_urls | jq > ../../$(APP_NAME)-urls.json
	@echo "wrote $(APP_NAME)-urls.json"

.PHONY: destroy
destroy: $(LAMBDA_BINARY)
	cd terraform/main && \
	terraform init $(TF_INIT_VARS) -reconfigure && \
	terraform destroy -auto-approve=$(AUTO_APPROVE) $(TF_VARS)
