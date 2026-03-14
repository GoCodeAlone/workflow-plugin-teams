package internal

import (
	"testing"
)

func TestNewTeamsModule_MissingCredentials(t *testing.T) {
	_, err := newTeamsModule("test", map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error creating module: %v", err)
	}
}

func TestTeamsModule_Init_MissingCredentials(t *testing.T) {
	m, _ := newTeamsModule("test", map[string]any{})
	err := m.Init()
	if err == nil {
		t.Fatal("expected error for missing credentials, got nil")
	}
}

func TestTeamsModule_Init_PartialCredentials(t *testing.T) {
	tests := []struct {
		name   string
		config map[string]any
	}{
		{"missing client_id and secret", map[string]any{"tenant_id": "t1"}},
		{"missing secret", map[string]any{"tenant_id": "t1", "client_id": "c1"}},
		{"missing tenant_id", map[string]any{"client_id": "c1", "client_secret": "s1"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, _ := newTeamsModule("test", tt.config)
			err := m.Init()
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tt.name)
			}
		})
	}
}

func TestGetModuleName_Default(t *testing.T) {
	name := getModuleName(map[string]any{})
	if name != "teams" {
		t.Errorf("expected 'teams', got %q", name)
	}
}

func TestGetModuleName_Custom(t *testing.T) {
	name := getModuleName(map[string]any{"module": "my-teams"})
	if name != "my-teams" {
		t.Errorf("expected 'my-teams', got %q", name)
	}
}

func TestResolveValue(t *testing.T) {
	current := map[string]any{"key": "from-current"}
	config := map[string]any{"key": "from-config"}
	if v := resolveValue("key", current, config); v != "from-current" {
		t.Errorf("expected 'from-current', got %q", v)
	}
	if v := resolveValue("key", map[string]any{}, config); v != "from-config" {
		t.Errorf("expected 'from-config', got %q", v)
	}
	if v := resolveValue("missing", map[string]any{}, map[string]any{}); v != "" {
		t.Errorf("expected '', got %q", v)
	}
}

func TestRegistry(t *testing.T) {
	// GetClient on missing name returns false
	_, ok := GetClient("nonexistent")
	if ok {
		t.Error("expected GetClient to return false for nonexistent client")
	}
	// UnregisterClient on missing name is safe
	UnregisterClient("nonexistent")
}

func TestStepRegistry_UnknownType(t *testing.T) {
	_, err := createStep("step.unknown_type", "test", map[string]any{})
	if err == nil {
		t.Fatal("expected error for unknown step type")
	}
}

func TestAllStepTypes(t *testing.T) {
	types := allStepTypes()
	if len(types) == 0 {
		t.Fatal("expected at least one step type")
	}
	expected := []string{
		"step.teams_send_message",
		"step.teams_send_card",
		"step.teams_reply_message",
		"step.teams_delete_message",
		"step.teams_create_channel",
		"step.teams_add_member",
		"step.teams_list_channel_messages",
		"step.teams_get_message",
	}
	typeSet := make(map[string]bool, len(types))
	for _, t := range types {
		typeSet[t] = true
	}
	for _, expected := range expected {
		if !typeSet[expected] {
			t.Errorf("missing step type %q", expected)
		}
	}
}

func TestStepExecute_MissingClient(t *testing.T) {
	steps := []struct {
		name    string
		stepFn  func(string, map[string]any) (interface{ Execute(interface{}, map[string]any, map[string]map[string]any, map[string]any, map[string]any, map[string]any) (interface{}, error) }, error)
	}{}
	_ = steps

	// Test each step returns an error output (not a hard error) when client is missing
	testCases := []struct {
		stepType string
		current  map[string]any
		config   map[string]any
	}{
		{
			"step.teams_send_message",
			map[string]any{"team_id": "t1", "channel_id": "c1", "content": "hello"},
			map[string]any{"module": "nonexistent"},
		},
		{
			"step.teams_send_card",
			map[string]any{"team_id": "t1", "channel_id": "c1", "card": `{"type":"AdaptiveCard"}`},
			map[string]any{"module": "nonexistent"},
		},
		{
			"step.teams_reply_message",
			map[string]any{"team_id": "t1", "channel_id": "c1", "message_id": "m1", "content": "reply"},
			map[string]any{"module": "nonexistent"},
		},
		{
			"step.teams_delete_message",
			map[string]any{"team_id": "t1", "channel_id": "c1", "message_id": "m1"},
			map[string]any{"module": "nonexistent"},
		},
		{
			"step.teams_create_channel",
			map[string]any{"team_id": "t1", "display_name": "new-channel"},
			map[string]any{"module": "nonexistent"},
		},
		{
			"step.teams_add_member",
			map[string]any{"team_id": "t1", "user_id": "u1"},
			map[string]any{"module": "nonexistent"},
		},
		{
			"step.teams_list_channel_messages",
			map[string]any{"team_id": "t1", "channel_id": "c1"},
			map[string]any{"module": "nonexistent"},
		},
		{
			"step.teams_get_message",
			map[string]any{"team_id": "t1", "channel_id": "c1", "message_id": "m1"},
			map[string]any{"module": "nonexistent"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.stepType, func(t *testing.T) {
			step, err := createStep(tc.stepType, "test", tc.config)
			if err != nil {
				t.Fatalf("createStep error: %v", err)
			}
			result, err := step.Execute(nil, nil, nil, tc.current, nil, tc.config)
			if err != nil {
				t.Fatalf("Execute returned hard error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if _, hasErr := result.Output["error"]; !hasErr {
				t.Errorf("expected output to contain 'error' key, got %v", result.Output)
			}
		})
	}
}

func TestStepExecute_MissingRequiredParams(t *testing.T) {
	testCases := []struct {
		stepType string
		current  map[string]any
		config   map[string]any
	}{
		{
			"step.teams_send_message",
			map[string]any{},
			map[string]any{},
		},
		{
			"step.teams_reply_message",
			map[string]any{"team_id": "t1"},
			map[string]any{},
		},
		{
			"step.teams_create_channel",
			map[string]any{"team_id": "t1"},
			map[string]any{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.stepType+"_missing_params", func(t *testing.T) {
			step, err := createStep(tc.stepType, "test", tc.config)
			if err != nil {
				t.Fatalf("createStep error: %v", err)
			}
			result, err := step.Execute(nil, nil, nil, tc.current, nil, tc.config)
			if err != nil {
				t.Fatalf("Execute returned hard error: %v", err)
			}
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if _, hasErr := result.Output["error"]; !hasErr {
				t.Errorf("expected output to contain 'error' key for missing params")
			}
		})
	}
}

func TestTeamsPlugin_Manifest(t *testing.T) {
	p := NewTeamsPlugin()
	m := p.Manifest()
	if m.Name == "" {
		t.Error("manifest name is empty")
	}
	if m.Version == "" {
		t.Error("manifest version is empty")
	}
}

func TestTeamsPlugin_Types(t *testing.T) {
	p := NewTeamsPlugin()

	tp, ok := p.(interface{ ModuleTypes() []string })
	if !ok {
		t.Fatal("plugin does not implement ModuleTypes()")
	}
	if len(tp.ModuleTypes()) == 0 {
		t.Error("expected at least one module type")
	}

	sp, ok := p.(interface{ StepTypes() []string })
	if !ok {
		t.Fatal("plugin does not implement StepTypes()")
	}
	if len(sp.StepTypes()) == 0 {
		t.Error("expected at least one step type")
	}

	trp, ok := p.(interface{ TriggerTypes() []string })
	if !ok {
		t.Fatal("plugin does not implement TriggerTypes()")
	}
	if len(trp.TriggerTypes()) == 0 {
		t.Error("expected at least one trigger type")
	}
}
