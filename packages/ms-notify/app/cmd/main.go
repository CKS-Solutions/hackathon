package main

import (
	"context"
	"net/http"

	http_internal "github.com/cks-solutions/hackathon/ms-notify/cmd/http"
	sqs_internal "github.com/cks-solutions/hackathon/ms-notify/cmd/sqs"
	"github.com/cks-solutions/hackathon/ms-notify/pkg/utils"
)

func main() {
	region := utils.GetRegion()
	stage := utils.GetStage()

	ctx := context.TODO()

	router := http_internal.NewRouter(ctx, region, stage)
	consumer := sqs_internal.NewSQSConsumer(ctx, region, stage)

	go consumer.Start()

	if err := http.ListenAndServe(":8080", router); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
