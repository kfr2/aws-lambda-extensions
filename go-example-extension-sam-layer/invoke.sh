#!/bin/sh

# Invoke the lambda and return its logs.
aws lambda invoke \
 --function-name "kfr2-extension-experiments-HelloWorldFunction-XAustUz2OOTo" \
 --payload '{"payload": "hello"}' response.json \
 --cli-binary-format raw-in-base64-out \
 --log-type Tail \
 --region us-east-1 \
 --profile sso-playground | jq -r .LogResult | base64 -D

printf "\n=====\n"

cat response.json | jq -r .env | base64 -D | jq
cat response.json | jq -r .secretsFile | base64 -D
