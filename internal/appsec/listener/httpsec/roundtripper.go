// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package httpsec

import (
	"github.com/0angelic0/dd-trace-go/internal/appsec/config"
	"github.com/0angelic0/dd-trace-go/internal/appsec/dyngo"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/httpsec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/waf"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/waf/addresses"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener"
)

type SSRFProtectionFeature struct{}

func (*SSRFProtectionFeature) String() string {
	return "SSRF Protection"
}

func (*SSRFProtectionFeature) Stop() {}

func NewSSRFProtectionFeature(config *config.Config, rootOp dyngo.Operation) (listener.Feature, error) {
	if !config.RASP || !config.SupportedAddresses.AnyOf(addresses.ServerIoNetURLAddr) {
		return nil, nil
	}

	feature := &SSRFProtectionFeature{}
	dyngo.On(rootOp, feature.OnStart)
	return feature, nil
}

func (*SSRFProtectionFeature) OnStart(op *httpsec.RoundTripOperation, args httpsec.RoundTripOperationArgs) {
	dyngo.EmitData(op, waf.RunEvent{
		Operation:      op,
		RunAddressData: addresses.NewAddressesBuilder().WithURL(args.URL).Build(),
	})
}
