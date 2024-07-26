default: deploy

GIT_REV=$(shell git rev-parse --short HEAD)

AWS_ACCOUNT_ID=$(shell aws sts get-caller-identity --query Account --output text)

AWS_REGION?=us-east-1

LAMBDA_BUILD=CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build

LAMBDA_BINARY=dist/burrow-$(GIT_REV).zip

BUCKET_NAME?=terraform-$(AWS_ACCOUNT_ID)

.PHONY: lambda
lambda:
	$(LAMBDA_BUILD) -o dist/bootstrap ./cmd/lambda
	zip -j $(LAMBDA_BINARY) dist/bootstrap

.PHONY: clean
clean:
	rm -rf dist && mkdir dist

.PHONY: deploy
deploy:
	$(MAKE) clean
	$(MAKE) lambda
	cd terraform/deployments/prod && \
	terraform init \
		-backend-config=bucket=$(BUCKET_NAME) \
		-backend-config=key=states/burrow/terraform.tfstate \
		-backend-config=region=$(AWS_REGION) && \
	terraform apply -auto-approve \
		-var revision=$(GIT_REV) \
		-var lambda_zip_path=../../../$(LAMBDA_BINARY) \
		-var bucket_name=$(BUCKET_NAME) \
		-var bucket_key=$(LAMBDA_BINARY)

.PHONY: destroy
destroy:
	cd terraform/deployments/prod && \
	terraform init \
		-backend-config=bucket=$(BUCKET_NAME) \
		-backend-config=key=states/burrow/terraform.tfstate \
		-backend-config=region=$(AWS_REGION) && \
	terraform destroy -auto-approve \
		-var revision=$(GIT_REV) \
		-var lambda_zip_path=../../../$(LAMBDA_BINARY) \
		-var bucket_name=$(BUCKET_NAME) \
		-var bucket_key=$(LAMBDA_BINARY)
