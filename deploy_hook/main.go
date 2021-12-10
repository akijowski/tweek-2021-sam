package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy"
	"github.com/aws/aws-sdk-go-v2/service/codedeploy/types"
	"log"
)

var codeDeployClient *codedeploy.Client

// DeploymentHook encapsulates the payload that is sent by AWS CodeDeploy when running pre/post traffic hooks.
//
// This is all you get.
type DeploymentHook struct {
	DeploymentID                  string `json:"DeploymentId"`
	LifecycleEventHookExecutionID string `json:"LifecycleEventHookExecutionId"`
}

func handler(ctx context.Context, event map[string]interface{}) error {
	log.Printf("event: %+v", event)

	var deploymentID string
	var executionId string
	if d, ok := event["DeploymentId"]; ok {
		deploymentID = d.(string)
	}
	if l, ok := event["LifecycleEventHookExecutionId"]; ok {
		executionId = l.(string)
	}
	log.Printf("found DeploymentId=%q and ExecutionId=%q", deploymentID, executionId)
	log.Print("automatically succeeding")

	_, err := codeDeployClient.PutLifecycleEventHookExecutionStatus(ctx, &codedeploy.PutLifecycleEventHookExecutionStatusInput{
		DeploymentId:                  aws.String(deploymentID),
		Status:                        types.LifecycleEventStatusSucceeded,
		LifecycleEventHookExecutionId: aws.String(executionId),
	})
	return err
}

func main() {
	lambda.Start(handler)
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		panic(err)
	}
	codeDeployClient = codedeploy.NewFromConfig(cfg)
}
