// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: MIT-0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"aws-lambda-extensions/go-example-extension/extension"
)

var (
	extensionName   = filepath.Base(os.Args[0]) // extension name has to match the filename
	extensionClient = extension.NewClient(os.Getenv("AWS_LAMBDA_RUNTIME_API"))
	printPrefix     = fmt.Sprintf("[%s]", extensionName)
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		s := <-sigs
		cancel()
		println(printPrefix, "Received", s)
		println(printPrefix, "Exiting")
	}()

	res, err := extensionClient.Register(ctx, extensionName)
	if err != nil {
		panic(err)
	}
	println(printPrefix, "Register response:", prettyPrint(res))

	// Retrieve the values for environment variables which begin with "SECRET_"
	// and store them in /tmp/variables. Afterwards, configure a timer to refresh
	// this file every minute.
	storeSecretEnvironmentVariables("devops/keys/lambda-extensions-test", "us-east-1")
	go func() {
		for now := range time.Tick(time.Minute) {
			fmt.Println(now, storeSecretEnvironmentVariables("devops/keys/lambda-extensions-test", "us-east-1"))
		}
	}()

	// Will block until shutdown event is received or cancelled via the context.
	processEvents(ctx)
}

func storeSecretEnvironmentVariables(secretName string, region string) string {
	fmt.Println(printPrefix, "Storing secret environment variables...")

	loadSecretsFromSecretsManager(secretName, region)

	f, err := os.Create("/tmp/variables")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "SECRET_") {
			println(printPrefix, "Found env variable to convert: ", e)
			envVar, _, _ := strings.Cut(e, "=")
			envVar = strings.ReplaceAll(envVar, "SECRET_", "")
			value := getSecretEnvironmentVariableValue(envVar)
			_, err := f.WriteString(fmt.Sprintf("%s=%s\n", envVar, value))
			if err != nil {
				println(printPrefix, "Error writing env variable to file: ", err)
			}
		}
	}

	return "Secrets have been written to /tmp/variables"
}

func processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			println(printPrefix, "Waiting for event...")
			res, err := extensionClient.NextEvent(ctx)
			if err != nil {
				println(printPrefix, "Error:", err)
				println(printPrefix, "Exiting")
				return
			}
			println(printPrefix, "Received event:", prettyPrint(res))
			// Exit if we receive a SHUTDOWN event
			if res.EventType == extension.Shutdown {
				println(printPrefix, "Received SHUTDOWN event")
				println(printPrefix, "Exiting")
				return
			}
		}
	}
}

func prettyPrint(v interface{}) string {
	data, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return ""
	}
	return string(data)
}
