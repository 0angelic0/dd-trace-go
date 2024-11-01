// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package sqlsec

import (
	"github.com/0angelic0/dd-trace-go/internal/appsec/config"
	"github.com/0angelic0/dd-trace-go/internal/appsec/dyngo"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/sqlsec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/waf"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/waf/addresses"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener"
)

type Feature struct{}

func (*Feature) String() string {
	return "SQLi Protection"
}

func (*Feature) Stop() {}

func NewSQLSecFeature(cfg *config.Config, rootOp dyngo.Operation) (listener.Feature, error) {
	if !cfg.RASP || !cfg.SupportedAddresses.AnyOf(addresses.ServerDBTypeAddr, addresses.ServerDBStatementAddr) {
		return nil, nil
	}

	feature := &Feature{}
	dyngo.On(rootOp, feature.OnStart)
	return feature, nil
}

func (*Feature) OnStart(op *sqlsec.SQLOperation, args sqlsec.SQLOperationArgs) {
	dyngo.EmitData(op, waf.RunEvent{
		Operation: op,
		RunAddressData: addresses.NewAddressesBuilder().
			WithDBStatement(args.Query).
			WithDBType(args.Driver).
			Build(),
	})
}
