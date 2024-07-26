default: deploy

APP_NAME?=burrow

GIT_REVISION=$(shell git rev-parse --short HEAD)

AWS_ACCOUNT_ID=$(shell aws sts get-caller-identity --query Account --output text)

AWS_REGION?=us-east-1

LAMBDA_BUILD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

LAMBDA_BINARY=dist/burrow-$(GIT_REVISION).zip

BUCKET_NAME?=terraform-$(AWS_ACCOUNT_ID)

AUTO_APPROVE?=false

$(LAMBDA_BINARY): $(shell find . -name '*.go') go.mod go.sum
	mkdir -p dist
	$(LAMBDA_BUILD) -o dist/bootstrap ./cmd/lambda
	zip -j $(LAMBDA_BINARY) dist/bootstrap

.PHONY: clean
clean:
	rm -rf dist

TF_INIT_VARS=-backend-config=bucket=$(BUCKET_NAME) \
	-backend-config=key=states/burrow/terraform.tfstate \
	-backend-config=region=$(AWS_REGION)

TF_VARS=-var name=$(APP_NAME) \
	-var git_revision=$(GIT_REVISION) \
	-var lambda_filename=../../$(LAMBDA_BINARY) \
	-var lambda_handler=burrow

.PHONY: deploy
deploy: $(LAMBDA_BINARY)
	cd terraform/main && \
	terraform init $(TF_INIT_VARS) && \
	terraform apply -auto-approve=$(AUTO_APPROVE) $(TF_VARS) && \
	terraform output -json function_urls | jq > ../../function_urls.json
	@echo "wrote function_urls.json"

.PHONY: destroy
destroy:
	cd terraform/main && \
	terraform init $(TF_INIT_VARS) && \
	terraform destroy -auto-approve=$(AUTO_APPROVE) $(TF_VARS)
