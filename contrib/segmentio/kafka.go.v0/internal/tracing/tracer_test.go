// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package tracing

import (
	"math"
	"testing"

	"github.com/0angelic0/dd-trace-go/internal/globalconfig"

	"github.com/stretchr/testify/assert"
)

func TestAnalyticsSettings(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		cfg := NewTracer(KafkaConfig{})
		assert.True(t, math.IsNaN(cfg.analyticsRate))
	})

	t.Run("global", func(t *testing.T) {
		t.Skip("global flag disabled")
		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		cfg := NewTracer(KafkaConfig{})
		assert.Equal(t, 0.4, cfg.analyticsRate)
	})

	t.Run("enabled", func(t *testing.T) {
		cfg := NewTracer(KafkaConfig{}, WithAnalytics(true))
		assert.Equal(t, 1.0, cfg.analyticsRate)
	})

	t.Run("override", func(t *testing.T) {
		rate := globalconfig.AnalyticsRate()
		defer globalconfig.SetAnalyticsRate(rate)
		globalconfig.SetAnalyticsRate(0.4)

		cfg := NewTracer(KafkaConfig{}, WithAnalyticsRate(0.2))
		assert.Equal(t, 0.2, cfg.analyticsRate)
	})

	t.Run("withEnv", func(t *testing.T) {
		t.Setenv("DD_DATA_STREAMS_ENABLED", "true")
		cfg := NewTracer(KafkaConfig{})
		assert.True(t, cfg.dataStreamsEnabled)
	})

	t.Run("optionOverridesEnv", func(t *testing.T) {
		t.Setenv("DD_DATA_STREAMS_ENABLED", "false")
		cfg := NewTracer(KafkaConfig{})
		WithDataStreams()(cfg)
		assert.True(t, cfg.dataStreamsEnabled)
	})
}
