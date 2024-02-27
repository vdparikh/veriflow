package main

import (
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/vdparikh/veriflow/veriflow"
)

func main() {
	app := veriflow.New()

	app.Logger.Info("Starting...")

	router := app.SetupRouter()

	// If running in AWS Lambda
	if lambdaRunningMode() {
		ginLambda := ginadapter.New(router)
		lambda.Start(func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return ginLambda.Proxy(req)
		})
	} else {
		// Start Local
		http.ListenAndServe(":8080", router)
	}
}

func lambdaRunningMode() bool {
	return os.Getenv("LAMBDA_TASK_ROOT") != ""
}
