// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package tracing

import (
	"net/http"
	"strings"

	"github.com/0angelic0/dd-trace-go/contrib/internal/httptrace"
	"github.com/0angelic0/dd-trace-go/contrib/internal/options"
	"github.com/0angelic0/dd-trace-go/ddtrace/tracer"
	"github.com/0angelic0/dd-trace-go/internal/telemetry"
)

const componentName = "julienschmidt/httprouter"

func init() {
	telemetry.LoadIntegration(componentName)
	tracer.MarkIntegrationImported("github.com/julienschmidt/httprouter")
}

type Router interface {
	Lookup(method string, path string) (any, []Param, bool)
}

type Param interface {
	GetKey() string
	GetValue() string
}

// BeforeHandle is an adapter of httptrace.BeforeHandle for julienschmidt/httprouter types.
func BeforeHandle[T any, WT Router](
	cfg *Config,
	router T,
	wrapRouter func(T) WT,
	w http.ResponseWriter,
	req *http.Request,
) (http.ResponseWriter, *http.Request, func(), bool) {
	wRouter := wrapRouter(router)
	// get the resource associated to this request
	route := req.URL.Path
	_, ps, _ := wRouter.Lookup(req.Method, route)
	for _, param := range ps {
		route = strings.Replace(route, param.GetValue(), ":"+param.GetKey(), 1)
	}

	resource := req.Method + " " + route
	spanOpts := options.Copy(cfg.spanOpts...) // spanOpts must be a copy of r.config.spanOpts, locally scoped, to avoid races.
	spanOpts = append(spanOpts, httptrace.HeaderTagsFromRequest(req, cfg.headerTags))

	serveCfg := &httptrace.ServeConfig{
		Service:  cfg.serviceName,
		Resource: resource,
		SpanOpts: spanOpts,
		Route:    route,
	}
	return httptrace.BeforeHandle(serveCfg, w, req)
}
