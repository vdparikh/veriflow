#!/bin/bash

# aws dynamodb delete-table --table-name VeriflowUsers --endpoint-url http://localhost:8000
aws dynamodb create-table \
    --table-name VeriflowUsers \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=email,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
        AttributeName=email,KeyType=RANGE \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --table-class STANDARD --endpoint-url http://localhost:8000

aws dynamodb update-table \
    --table-name VeriflowUsers \
    --attribute-definitions AttributeName=email,AttributeType=S \
    --global-secondary-index-updates \
        "[{\"Create\":{\"IndexName\":\"EmailIndex\",\"KeySchema\":[{\"AttributeName\":\"email\",\"KeyType\":\"HASH\"}],\"Projection\":{\"ProjectionType\":\"ALL\"},\"ProvisionedThroughput\":{\"ReadCapacityUnits\":5,\"WriteCapacityUnits\":5}}}]" \
    --endpoint-url http://localhost:8000

#aws dynamodb delete-table --table-name VeriflowRequests --endpoint-url http://localhost:8000
aws dynamodb create-table \
    --table-name VeriflowRequests \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
        AttributeName=requestor_email,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
        AttributeName=requestor_email,KeyType=RANGE \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --table-class STANDARD --endpoint-url http://localhost:8000

aws dynamodb update-table \
    --table-name VeriflowRequests \
    --attribute-definitions AttributeName=email,AttributeType=S \
    --global-secondary-index-updates \
        "[{\"Create\":{\"IndexName\":\"EmailIndex\",\"KeySchema\":[{\"AttributeName\":\"requestor_email\",\"KeyType\":\"HASH\"}],\"Projection\":{\"ProjectionType\":\"ALL\"},\"ProvisionedThroughput\":{\"ReadCapacityUnits\":5,\"WriteCapacityUnits\":5}}}]" \
    --endpoint-url http://localhost:8000        

#aws dynamodb delete-table --table-name VeriflowWebAuthnSessions --endpoint-url http://localhost:8000
aws dynamodb create-table \
    --table-name VeriflowWebAuthnSessions \
    --attribute-definitions \
        AttributeName=id,AttributeType=S \
    --key-schema \
        AttributeName=id,KeyType=HASH \
    --provisioned-throughput \
        ReadCapacityUnits=5,WriteCapacityUnits=5 \
    --table-class STANDARD --endpoint-url http://localhost:8000    


# Verify if tables are created
aws dynamodb list-tables --endpoint-url http://localhost:8000    