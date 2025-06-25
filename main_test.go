package main

import (
	"context"
	"flag"
	"os"
	"testing"

	"github.com/ONSdigital/dis-redirect-api/features/steps"
	componentTest "github.com/ONSdigital/dp-component-test"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var componentFlag = flag.Bool("component", false, "perform component tests")

type ComponentTest struct {
	RedisFeature *componentTest.RedisFeature
}

func (f *ComponentTest) InitializeScenario(ctx *godog.ScenarioContext) {
	f.RedisFeature = componentTest.NewRedisFeature()
	redirectAPIComponent, err := steps.NewRedirectComponent(f.RedisFeature)
	if err != nil {
		log.Error(context.Background(), "failed to create redirect api component", err)
		os.Exit(1)
	}

	apiFeature := redirectAPIComponent.InitAPIFeature()

	ctx.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		if f.RedisFeature == nil {
			f.RedisFeature = componentTest.NewRedisFeature()
		}
		apiFeature.Reset()

		return ctx, nil
	})

	ctx.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		if closeErr := f.RedisFeature.Close(); closeErr != nil {
			log.Error(context.Background(), "error occured while closing the RedisFeature", closeErr)
			os.Exit(1)
		}

		apiFeature.Reset()

		return ctx, nil
	})

	f.RedisFeature.RegisterSteps(ctx)
	redirectAPIComponent.RegisterSteps(ctx)
}

func (f *ComponentTest) InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
	})

	ctx.AfterSuite(func() {
	})
}

func TestComponent(t *testing.T) {
	if *componentFlag {
		status := 0

		var opts = godog.Options{
			Output: colors.Colored(os.Stdout),
			Format: "pretty",
			Paths:  flag.Args(),
			Strict: true,
		}

		f := &ComponentTest{}

		status = godog.TestSuite{
			Name:                 "feature_tests",
			ScenarioInitializer:  f.InitializeScenario,
			TestSuiteInitializer: f.InitializeTestSuite,
			Options:              &opts,
		}.Run()

		if status > 0 {
			t.Fail()
		}
	} else {
		t.Skip("component flag required to run component tests")
	}
}
