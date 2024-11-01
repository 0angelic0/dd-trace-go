// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package restful_test

import (
	"io"
	"log"
	"net/http"

	restfultrace "github.com/0angelic0/dd-trace-go/contrib/emicklei/go-restful.v3"
	"github.com/0angelic0/dd-trace-go/ddtrace/tracer"

	"github.com/emicklei/go-restful/v3"
)

// To start tracing requests, add the trace filter to your go-restful router.
func Example() {
	// create new go-restful service
	ws := new(restful.WebService)

	// create the Datadog filter
	filter := restfultrace.FilterFunc(
		restfultrace.WithServiceName("my-service"),
	)

	// use it
	ws.Filter(filter)

	// set endpoint
	ws.Route(ws.GET("/hello").To(
		func(request *restful.Request, response *restful.Response) {
			io.WriteString(response, "world")
		}))
	restful.Add(ws)

	// serve request
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func Example_spanFromContext() {
	ws := new(restful.WebService)
	ws.Filter(restfultrace.FilterFunc(
		restfultrace.WithServiceName("my-service"),
	))

	ws.Route(ws.GET("/image/encode").To(
		func(request *restful.Request, response *restful.Response) {
			// create a child span to track operation timing.
			encodeSpan, _ := tracer.StartSpanFromContext(request.Request.Context(), "image.encode")
			// encode a image
			encodeSpan.Finish()
		}))
}
