// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package usersec

import (
	"github.com/0angelic0/dd-trace-go/internal/appsec/config"
	"github.com/0angelic0/dd-trace-go/internal/appsec/dyngo"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/usersec"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/waf"
	"github.com/0angelic0/dd-trace-go/internal/appsec/emitter/waf/addresses"
	"github.com/0angelic0/dd-trace-go/internal/appsec/listener"
)

type Feature struct{}

func (*Feature) String() string {
	return "User Security"
}

func (*Feature) Stop() {}

func NewUserSecFeature(cfg *config.Config, rootOp dyngo.Operation) (listener.Feature, error) {
	if !cfg.SupportedAddresses.AnyOf(
		addresses.UserIDAddr,
		addresses.UserSessionIDAddr,
		addresses.UserLoginSuccessAddr,
		addresses.UserLoginFailureAddr) {
		return nil, nil
	}

	feature := &Feature{}
	dyngo.OnFinish(rootOp, feature.OnFinish)
	return feature, nil
}

func (*Feature) OnFinish(op *usersec.UserLoginOperation, res usersec.UserLoginOperationRes) {
	builder := addresses.NewAddressesBuilder()

	switch op.EventType {
	case usersec.UserLoginSuccess:
		builder = builder.WithUserLoginSuccess().
			WithUserID(res.UserID).
			WithUserSessionID(res.SessionID)
	case usersec.UserLoginFailure:
		builder = builder.WithUserLoginFailure().
			WithUserID(res.UserID)
	case usersec.UserSet:
		builder = builder.WithUserID(res.UserID).
			WithUserSessionID(res.SessionID)
	}

	dyngo.EmitData(op, waf.RunEvent{
		Operation:      op,
		RunAddressData: builder.Build(),
	})
}
