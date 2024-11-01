// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package appsec

import (
	"errors"

	"github.com/0angelic0/dd-trace-go/internal/appsec/dyngo"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener/graphqlsec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener/grpcsec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener/httpsec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener/ossec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener/sqlsec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener/trace"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener/usersec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener/waf"
	"github.com/0angelic0/dd-trace-go/internal/log"
)

var features = []listener.NewFeature{
	trace.NewAppsecSpanTransport,
	waf.NewWAFFeature,
	httpsec.NewHTTPSecFeature,
	grpcsec.NewGRPCSecFeature,
	graphqlsec.NewGraphQLSecFeature,
	usersec.NewUserSecFeature,
	sqlsec.NewSQLSecFeature,
	ossec.NewOSSecFeature,
	httpsec.NewSSRFProtectionFeature,
}

func (a *appsec) SwapRootOperation() error {
	newRoot := dyngo.NewRootOperation()
	newFeatures := make([]listener.Feature, 0, len(features))
	var featureErrors []error
	for _, newFeature := range features {
		feature, err := newFeature(a.cfg, newRoot)
		if err != nil {
			featureErrors = append(featureErrors, err)
			continue
		}

		// If error is nil and feature is nil, it means the feature did not activate itself
		if feature == nil {
			continue
		}

		newFeatures = append(newFeatures, feature)
	}

	err := errors.Join(featureErrors...)
	if err != nil {
		for _, feature := range newFeatures {
			feature.Stop()
		}
		return err
	}

	a.featuresMu.Lock()
	defer a.featuresMu.Unlock()

	oldFeatures := a.features
	a.features = newFeatures

	log.Debug("appsec: stopping the following features: %v", oldFeatures)
	log.Debug("appsec: starting the following features: %v", newFeatures)

	dyngo.SwapRootOperation(newRoot)

	log.Debug("appsec: swapped root operation")

	for _, oldFeature := range oldFeatures {
		oldFeature.Stop()
	}

	return nil
}
