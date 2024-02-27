.PHONY: build package deploy config

S3_BUCKET ?= $(VERIFLOW_BUCKET)
STACK_NAME = veriflow
BUILD_OUTPUT = bootstrap
ZIP_FILE = bootstrap.zip
MAIN_GO = cmd/main.go
TEMPLATES_DIR = templates
CONFIG_FILE = config.yaml

all: build package deploy

build:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BUILD_OUTPUT) $(MAIN_GO)
	cp -r $(TEMPLATES_DIR) bin/
	cd bin && zip -FS $(ZIP_FILE) bootstrap $(TEMPLATES_DIR)/*

package:
	sam package --template-file template.yaml --s3-bucket $(S3_BUCKET) --output-template-file packaged.yaml --force-upload

deploy:
	sam deploy --template-file packaged.yaml --stack-name $(STACK_NAME) --capabilities CAPABILITY_NAMED_IAM --parameter-overrides VeriflowBucket=$(S3_BUCKET)

config:
	aws s3 cp $(CONFIG_FILE) s3://$(S3_BUCKET)/$(CONFIG_FILE)

clean:
	rm -f bin/$(BUILD_OUTPUT)
	rm -rf bin/$(TEMPLATES_DIR)
	rm -f bin/$(ZIP_FILE)

local:
	CGO_ENABLED=0 GOOS=linux go build -o main -v cmd/main.go
	docker-compose build
	docker-compose up