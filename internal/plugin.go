// Package internal implements the workflow-plugin-teams plugin.
package internal

import (
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// teamsPlugin implements sdk.PluginProvider, sdk.ModuleProvider, sdk.StepProvider, and sdk.TriggerProvider.
type teamsPlugin struct{}

// NewTeamsPlugin returns a new teamsPlugin instance.
func NewTeamsPlugin() sdk.PluginProvider {
	return &teamsPlugin{}
}

// Manifest returns plugin metadata.
func (p *teamsPlugin) Manifest() sdk.PluginManifest {
	return sdk.PluginManifest{
		Name:        "teams",
		Version:     "0.1.0",
		Author:      "GoCodeAlone",
		Description: "Microsoft Teams messaging plugin via Graph API",
	}
}

// ModuleTypes returns the module type names this plugin provides.
func (p *teamsPlugin) ModuleTypes() []string {
	return []string{"teams.provider"}
}

// CreateModule creates a module instance of the given type.
func (p *teamsPlugin) CreateModule(typeName, name string, config map[string]any) (sdk.ModuleInstance, error) {
	switch typeName {
	case "teams.provider":
		return newTeamsModule(name, config)
	default:
		return nil, fmt.Errorf("teams plugin: unknown module type %q", typeName)
	}
}

// StepTypes returns the step type names this plugin provides.
func (p *teamsPlugin) StepTypes() []string {
	return allStepTypes()
}

// CreateStep creates a step instance of the given type.
func (p *teamsPlugin) CreateStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	return createStep(typeName, name, config)
}

// TriggerTypes returns the trigger type names this plugin provides.
func (p *teamsPlugin) TriggerTypes() []string {
	return []string{"trigger.teams"}
}

// CreateTrigger creates a trigger instance of the given type.
func (p *teamsPlugin) CreateTrigger(typeName string, config map[string]any, callback sdk.TriggerCallback) (sdk.TriggerInstance, error) {
	switch typeName {
	case "trigger.teams":
		return newTeamsTrigger(typeName, config, callback)
	default:
		return nil, fmt.Errorf("teams plugin: unknown trigger type %q", typeName)
	}
}
