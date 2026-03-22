package internal

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// noopTriggerCallback is a do-nothing callback used in trigger tests.
func noopTriggerCallback(_ string, _ map[string]any) error { return nil }

// --- Step 1: Regression test proving the vulnerability ---
//
// Before the fix, this test FAILS: a forged notification is processed (202)
// because handleNotification never checks clientState. After the fix, the
// test PASSES: the forged request is rejected (403).

func TestHandleNotification_ForgedClientState_Rejected(t *testing.T) {
	trigger := &teamsTrigger{
		name:        "test",
		clientState: "real-secret",
		callback:    noopTriggerCallback,
	}

	payload := changeNotificationCollection{
		Value: []changeNotification{
			{ChangeType: "created", Resource: "test/resource", ClientState: "attacker-forged"},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/teams/notifications", bytes.NewReader(body))
	w := httptest.NewRecorder()
	trigger.handleNotification(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("forged clientState should be rejected with 403, got %d (vulnerability present)", w.Code)
	}
}

// --- Step 2: Tests for the fixed behaviour ---

func TestHandleNotification_EmptyClientState_Rejected(t *testing.T) {
	trigger := &teamsTrigger{
		name:        "test",
		clientState: "real-secret",
		callback:    noopTriggerCallback,
	}

	payload := changeNotificationCollection{
		Value: []changeNotification{
			{ChangeType: "created", Resource: "test/resource", ClientState: ""},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/teams/notifications", bytes.NewReader(body))
	w := httptest.NewRecorder()
	trigger.handleNotification(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("empty clientState should be rejected with 403, got %d", w.Code)
	}
}

func TestHandleNotification_CorrectClientState_Accepted(t *testing.T) {
	trigger := &teamsTrigger{
		name:        "test",
		clientState: "real-secret",
		callback:    noopTriggerCallback,
	}

	payload := changeNotificationCollection{
		Value: []changeNotification{
			{ChangeType: "created", Resource: "test/resource", ClientState: "real-secret"},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/teams/notifications", bytes.NewReader(body))
	w := httptest.NewRecorder()
	trigger.handleNotification(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("correct clientState should be accepted with 202, got %d", w.Code)
	}
}

// TestHandleNotification_NoClientStateConfigured_PassThrough verifies that when
// no clientState is set on the trigger (e.g. subscription was created externally),
// notifications pass through unchanged to preserve backwards compatibility.
func TestHandleNotification_NoClientStateConfigured_PassThrough(t *testing.T) {
	trigger := &teamsTrigger{
		name:        "test",
		clientState: "",
		callback:    noopTriggerCallback,
	}

	payload := changeNotificationCollection{
		Value: []changeNotification{
			{ChangeType: "created", Resource: "test/resource", ClientState: "anything"},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/teams/notifications", bytes.NewReader(body))
	w := httptest.NewRecorder()
	trigger.handleNotification(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202 when no clientState is configured on trigger, got %d", w.Code)
	}
}

// TestHandleNotification_MultiNotification_OneForged verifies that a batch
// containing even one notification with a wrong clientState rejects the whole request.
func TestHandleNotification_MultiNotification_OneForged(t *testing.T) {
	trigger := &teamsTrigger{
		name:        "test",
		clientState: "real-secret",
		callback:    noopTriggerCallback,
	}

	payload := changeNotificationCollection{
		Value: []changeNotification{
			{ChangeType: "created", Resource: "test/resource", ClientState: "real-secret"},
			{ChangeType: "updated", Resource: "test/resource", ClientState: "attacker-forged"},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/teams/notifications", bytes.NewReader(body))
	w := httptest.NewRecorder()
	trigger.handleNotification(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("batch with one forged clientState should be rejected with 403, got %d", w.Code)
	}
}

// TestHandleNotification_ValidationToken_NotAffectedByClientStateCheck verifies
// that the Graph subscription validation handshake is never gated by clientState.
func TestHandleNotification_ValidationToken_NotAffectedByClientStateCheck(t *testing.T) {
	trigger := &teamsTrigger{
		name:        "test",
		clientState: "real-secret",
		callback:    noopTriggerCallback,
	}

	req := httptest.NewRequest(http.MethodGet, "/teams/notifications?validationToken=abc123", http.NoBody)
	w := httptest.NewRecorder()
	trigger.handleNotification(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("validation token handshake should return 200, got %d", w.Code)
	}
	if w.Body.String() != "abc123" {
		t.Fatalf("expected echoed validation token 'abc123', got %q", w.Body.String())
	}
}

// TestNewTeamsTrigger_ClientStateInitiallyEmpty verifies that a newly created
// trigger starts with an empty clientState (set only during createSubscription).
func TestNewTeamsTrigger_ClientStateInitiallyEmpty(t *testing.T) {
	trigger, err := newTeamsTrigger("test", map[string]any{
		"team_id":      "t1",
		"channel_id":   "c1",
		"callback_url": "https://example.com",
	}, noopTriggerCallback)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if trigger.clientState != "" {
		t.Errorf("expected empty clientState before subscription creation, got %q", trigger.clientState)
	}
}
