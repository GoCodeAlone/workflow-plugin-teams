package internal

import (
	"encoding/json"
	"os"
	"sort"
	"testing"
)

// TestPluginCapabilitiesMatchManifest verifies that the types registered by
// the Go plugin code exactly match the capabilities declared in plugin.json.
// This catches renames or removals in the code that would otherwise leave
// stale advertised metadata while CI continues to pass.
func TestPluginCapabilitiesMatchManifest(t *testing.T) {
	data, err := os.ReadFile("../plugin.json")
	if err != nil {
		t.Fatalf("read plugin.json: %v", err)
	}

	var manifest struct {
		Capabilities struct {
			ModuleTypes  []string `json:"moduleTypes"`
			StepTypes    []string `json:"stepTypes"`
			TriggerTypes []string `json:"triggerTypes"`
		} `json:"capabilities"`
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse plugin.json: %v", err)
	}

	p := NewTeamsPlugin()

	mp, ok := p.(interface{ ModuleTypes() []string })
	if !ok {
		t.Fatal("plugin does not implement ModuleTypes()")
	}
	checkTypesMatch(t, "capabilities.moduleTypes", manifest.Capabilities.ModuleTypes, mp.ModuleTypes())

	sp, ok := p.(interface{ StepTypes() []string })
	if !ok {
		t.Fatal("plugin does not implement StepTypes()")
	}
	checkTypesMatch(t, "capabilities.stepTypes", manifest.Capabilities.StepTypes, sp.StepTypes())

	trp, ok := p.(interface{ TriggerTypes() []string })
	if !ok {
		t.Fatal("plugin does not implement TriggerTypes()")
	}
	checkTypesMatch(t, "capabilities.triggerTypes", manifest.Capabilities.TriggerTypes, trp.TriggerTypes())
}

// checkTypesMatch asserts that wantFromManifest and gotFromCode contain the
// same strings (order-independent).  Both directions are checked so that
// additions in the manifest (missing from code) and additions in code
// (missing from manifest) both fail the test.
func checkTypesMatch(t *testing.T, field string, wantFromManifest, gotFromCode []string) {
	t.Helper()

	manifestSet := toSet(wantFromManifest)
	codeSet := toSet(gotFromCode)

	for _, typ := range sorted(wantFromManifest) {
		if !codeSet[typ] {
			t.Errorf("%s: manifest declares %q but plugin code does not register it", field, typ)
		}
	}
	for _, typ := range sorted(gotFromCode) {
		if !manifestSet[typ] {
			t.Errorf("%s: plugin code registers %q but manifest does not declare it", field, typ)
		}
	}
}

func toSet(ss []string) map[string]bool {
	m := make(map[string]bool, len(ss))
	for _, s := range ss {
		m[s] = true
	}
	return m
}

func sorted(ss []string) []string {
	out := make([]string, len(ss))
	copy(out, ss)
	sort.Strings(out)
	return out
}
