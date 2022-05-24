function handler () {
    EVENT_DATA=$1

    ENVIRONMENT=$(./jq -ncr '$ENV|@base64')
    SECRETS=$(base64 /tmp/variables)

    RESPONSE="{\"statusCode\": 200, \"body\": \"Hello from Lambda!\", \"env\": \"$ENVIRONMENT\", \"secretsFile\": \"$SECRETS\"}"
    echo $RESPONSE
}
