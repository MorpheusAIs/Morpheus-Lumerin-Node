package main

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/urfave/cli/v2"
)

// buildApp mirrors the wiring in main(): the API client is created in Before,
// so that help output does not depend on a reachable router.
//
// clientErr stands in for NewApiGatewayClientFromEnv failing, which is what
// happens when the router is down or the cookie cannot be read.
func buildApp(clientErr error, built *int) *cli.App {
	return &cli.App{
		Usage:  "A client to call the Morpheus Lumerin API",
		Writer: new(bytes.Buffer),
		Before: func(cCtx *cli.Context) error {
			switch cCtx.Args().First() {
			case "help", "h":
				return nil
			}
			*built++
			return clientErr
		},
		Commands: []*cli.Command{
			{
				Name:    "healthcheck",
				Aliases: []string{"he"},
				Usage:   "morpheus healthcheck",
				Action:  func(*cli.Context) error { return nil },
			},
			{
				Name:  "wallet",
				Usage: "morpheus wallet",
				Subcommands: []*cli.Command{
					{Name: "balance", Action: func(*cli.Context) error { return nil }},
				},
				Action: func(*cli.Context) error { return nil },
			},
		},
	}
}

// TestHelpWorksWithoutARouter is the regression test for the startup order:
// the client used to be built in main() before arg parsing, so every
// invocation -- including help -- failed when the router was unreachable.
func TestHelpWorksWithoutARouter(t *testing.T) {
	down := errors.New("can't get cookie path: connection refused")

	for _, args := range [][]string{
		{"mor-cli", "--help"},
		{"mor-cli", "-h"},
		{"mor-cli", "help"},
		{"mor-cli", "help", "wallet"},
		{"mor-cli", "help", "healthcheck"},
	} {
		t.Run(strings.Join(args[1:], " "), func(t *testing.T) {
			built := 0
			if err := buildApp(down, &built).Run(args); err != nil {
				t.Errorf("%v should succeed with no router, got: %v", args[1:], err)
			}
			if built != 0 {
				t.Errorf("%v built the API client %d time(s); help must not need one", args[1:], built)
			}
		})
	}
}

// TestCommandsStillRequireAClient: the fix must not let a real command run
// without a client, which would dereference nil.
func TestCommandsStillRequireAClient(t *testing.T) {
	down := errors.New("can't get cookie path: connection refused")

	for _, args := range [][]string{
		{"mor-cli", "healthcheck"},
		{"mor-cli", "he"},
		{"mor-cli", "wallet"},
		{"mor-cli", "wallet", "balance"},
	} {
		t.Run(strings.Join(args[1:], " "), func(t *testing.T) {
			built := 0
			err := buildApp(down, &built).Run(args)
			if err == nil {
				t.Errorf("%v must fail when the client cannot be built", args[1:])
			}
			if built != 1 {
				t.Errorf("%v built the client %d time(s), want 1", args[1:], built)
			}
		})
	}
}

// TestCommandsRunWhenTheClientBuilds covers the other direction: Before has to
// actually populate the client on the happy path.
func TestCommandsRunWhenTheClientBuilds(t *testing.T) {
	built := 0
	if err := buildApp(nil, &built).Run([]string{"mor-cli", "healthcheck"}); err != nil {
		t.Fatalf("command should run when the client builds: %v", err)
	}
	if built != 1 {
		t.Errorf("client built %d time(s), want 1", built)
	}
}
