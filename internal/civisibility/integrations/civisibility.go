// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2024 Datadog, Inc.

package integrations

import (
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"

	"github.com/0angelic0/dd-trace-go/ddtrace/mocktracer"
	"github.com/0angelic0/dd-trace-go/ddtrace/tracer"
	"github.com/0angelic0/dd-trace-go/internal"
	"github.com/0angelic0/dd-trace-go/internal/civisibility/constants"
	"github.com/0angelic0/dd-trace-go/internal/civisibility/utils"
	"github.com/0angelic0/dd-trace-go/internal/log"
)

// ciVisibilityCloseAction defines an action to be executed when CI visibility is closing.
type ciVisibilityCloseAction func()

var (
	// ciVisibilityInitializationOnce ensures we initialize the CI visibility tracer only once.
	ciVisibilityInitializationOnce sync.Once

	// closeActions holds CI visibility close actions.
	closeActions []ciVisibilityCloseAction

	// closeActionsMutex synchronizes access to closeActions.
	closeActionsMutex sync.Mutex

	// mTracer contains the mock tracer instance for testing purposes
	mTracer mocktracer.Tracer
)

// EnsureCiVisibilityInitialization initializes the CI visibility tracer if it hasn't been initialized already.
func EnsureCiVisibilityInitialization() {
	internalCiVisibilityInitialization(func(opts []tracer.StartOption) {
		// Initialize the tracer.
		tracer.Start(opts...)
	})
}

// InitializeCIVisibilityMock initialize the mocktracer for CI Visibility usage
func InitializeCIVisibilityMock() mocktracer.Tracer {
	internalCiVisibilityInitialization(func([]tracer.StartOption) {
		// Initialize the mocktracer
		mTracer = mocktracer.Start()
	})
	return mTracer
}

func internalCiVisibilityInitialization(tracerInitializer func([]tracer.StartOption)) {
	ciVisibilityInitializationOnce.Do(func() {
		// check the debug flag to enable debug logs. The tracer initialization happens
		// after the CI Visibility initialization so we need to handle this flag ourselves
		if internal.BoolEnv("DD_TRACE_DEBUG", false) {
			log.SetLevel(log.LevelDebug)
		}

		log.Debug("civisibility: initializing")

		// Since calling this method indicates we are in CI Visibility mode, set the environment variable.
		_ = os.Setenv(constants.CIVisibilityEnabledEnvironmentVariable, "1")

		// Avoid sampling rate warning (in CI Visibility mode we send all data)
		_ = os.Setenv("DD_TRACE_SAMPLE_RATE", "1")

		// Preload the CodeOwner file
		_ = utils.GetCodeOwners()

		// Preload all CI, Git, and CodeOwners tags.
		ciTags := utils.GetCITags()
		_ = utils.GetCIMetrics()

		// Check if DD_SERVICE has been set; otherwise default to the repo name (from the spec).
		var opts []tracer.StartOption
		serviceName := os.Getenv("DD_SERVICE")
		if serviceName == "" {
			if repoURL, ok := ciTags[constants.GitRepositoryURL]; ok {
				// regex to sanitize the repository url to be used as a service name
				repoRegex := regexp.MustCompile(`(?m)/([a-zA-Z0-9\-_.]*)$`)
				matches := repoRegex.FindStringSubmatch(repoURL)
				if len(matches) > 1 {
					repoURL = strings.TrimSuffix(matches[1], ".git")
				}
				serviceName = repoURL
				opts = append(opts, tracer.WithService(serviceName))
			}
		}

		// Initializing additional features asynchronously
		go func() { ensureAdditionalFeaturesInitialization(serviceName) }()

		// Initialize the tracer
		log.Debug("civisibility: initializing tracer")
		tracerInitializer(opts)

		// Handle SIGINT and SIGTERM signals to ensure we close all open spans and flush the tracer before exiting
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			<-signals
			ExitCiVisibility()
			os.Exit(1)
		}()
	})
}

// PushCiVisibilityCloseAction adds a close action to be executed when CI visibility exits.
func PushCiVisibilityCloseAction(action ciVisibilityCloseAction) {
	closeActionsMutex.Lock()
	defer closeActionsMutex.Unlock()
	closeActions = append([]ciVisibilityCloseAction{action}, closeActions...)
}

// ExitCiVisibility executes all registered close actions and stops the tracer.
func ExitCiVisibility() {
	log.Debug("civisibility: exiting")
	closeActionsMutex.Lock()
	defer closeActionsMutex.Unlock()
	defer func() {
		closeActions = []ciVisibilityCloseAction{}

		log.Debug("civisibility: flushing and stopping tracer")
		tracer.Flush()
		tracer.Stop()
		log.Debug("civisibility: done.")
	}()
	for _, v := range closeActions {
		v()
	}
}
