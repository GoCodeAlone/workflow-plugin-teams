// Package main is the entry point for the workflow-plugin-teams external plugin binary.
// The binary is started by the Workflow engine over gRPC via the GoCodeAlone/go-plugin host.
package main

import (
	"github.com/GoCodeAlone/workflow-plugin-teams/internal"
	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// version is set at build time via -ldflags "-X main.version=X.Y.Z".
var version = "0.0.0"

func main() {
	internal.Version = version
	sdk.Serve(internal.NewTeamsPlugin())
}
