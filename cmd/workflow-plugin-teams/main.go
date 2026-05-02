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
	// Only forward the ldflags version to internal.Version when the linker
	// actually injected a real release tag.  If a build sets only
	// -X github.com/GoCodeAlone/workflow-plugin-teams/internal.Version=X.Y.Z
	// without also setting -X main.version=X.Y.Z, the default "0.0.0" must
	// not clobber the already-injected value.
	if version != "0.0.0" {
		internal.Version = version
	}
	sdk.Serve(internal.NewTeamsPlugin())
}
