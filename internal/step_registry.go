package internal

import (
	"fmt"

	sdk "github.com/GoCodeAlone/workflow/plugin/external/sdk"
)

// stepConstructor is a function that creates a StepInstance.
type stepConstructor func(name string, config map[string]any) (sdk.StepInstance, error)

// stepRegistry maps step type strings to constructor functions.
var stepRegistry = map[string]stepConstructor{
	"step.teams_send_message":           func(n string, c map[string]any) (sdk.StepInstance, error) { return newSendMessageStep(n, c) },
	"step.teams_send_card":              func(n string, c map[string]any) (sdk.StepInstance, error) { return newSendCardStep(n, c) },
	"step.teams_reply_message":          func(n string, c map[string]any) (sdk.StepInstance, error) { return newReplyMessageStep(n, c) },
	"step.teams_delete_message":         func(n string, c map[string]any) (sdk.StepInstance, error) { return newDeleteMessageStep(n, c) },
	"step.teams_create_channel":         func(n string, c map[string]any) (sdk.StepInstance, error) { return newCreateChannelStep(n, c) },
	"step.teams_add_member":             func(n string, c map[string]any) (sdk.StepInstance, error) { return newAddMemberStep(n, c) },
	"step.teams_list_channel_messages":  func(n string, c map[string]any) (sdk.StepInstance, error) { return newListChannelMessagesStep(n, c) },
	"step.teams_get_message":            func(n string, c map[string]any) (sdk.StepInstance, error) { return newGetMessageStep(n, c) },
}

// createStep dispatches to the appropriate step constructor.
func createStep(typeName, name string, config map[string]any) (sdk.StepInstance, error) {
	constructor, ok := stepRegistry[typeName]
	if !ok {
		return nil, fmt.Errorf("teams plugin: unknown step type %q", typeName)
	}
	return constructor(name, config)
}

// allStepTypes returns all registered step type strings.
func allStepTypes() []string {
	types := make([]string, 0, len(stepRegistry))
	for k := range stepRegistry {
		types = append(types, k)
	}
	return types
}
