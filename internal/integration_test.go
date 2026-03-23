package internal_test

import (
	"testing"

	"github.com/GoCodeAlone/workflow/wftest"
)

func TestIntegration_SendMessage(t *testing.T) {
	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  notify:
    steps:
      - name: send
        type: step.teams_send_message
        config:
          team_id: "team123"
          channel_id: "chan456"
          content: "hello world"
      - name: confirm
        type: step.set
        config:
          values:
            sent: true
`),
		wftest.MockStep("step.teams_send_message", wftest.Returns(map[string]any{
			"message_id": "msg001",
			"id":         "msg001",
			"web_url":    "https://teams.microsoft.com/msg/001",
			"team_id":    "team123",
			"channel_id": "chan456",
		})),
	)

	result := h.ExecutePipeline("notify", nil)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.Output["sent"] != true {
		t.Errorf("expected sent=true, got %v", result.Output["sent"])
	}

	sendOut := result.StepOutput("send")
	if sendOut["message_id"] != "msg001" {
		t.Errorf("expected message_id=msg001, got %v", sendOut["message_id"])
	}
}

func TestIntegration_ReplyMessage(t *testing.T) {
	rec := wftest.RecordStep("step.teams_reply_message")
	rec.WithOutput(map[string]any{
		"reply_id":    "reply001",
		"message_id":  "msg001",
		"team_id":     "team123",
		"channel_id":  "chan456",
		"reply_to_id": "msg001",
	})

	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  reply_flow:
    steps:
      - name: reply
        type: step.teams_reply_message
        config:
          team_id: "team123"
          channel_id: "chan456"
          message_id: "msg001"
          content: "this is a reply"
      - name: done
        type: step.set
        config:
          values:
            replied: true
`),
		rec,
	)

	result := h.ExecutePipeline("reply_flow", nil)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.Output["replied"] != true {
		t.Errorf("expected replied=true, got %v", result.Output["replied"])
	}
	if rec.CallCount() != 1 {
		t.Errorf("expected reply step called once, got %d", rec.CallCount())
	}

	calls := rec.Calls()
	if calls[0].Config["message_id"] != "msg001" {
		t.Errorf("expected config message_id=msg001, got %v", calls[0].Config["message_id"])
	}
}

func TestIntegration_CreateChannelAndSendMessage(t *testing.T) {
	h := wftest.New(t, wftest.WithYAML(`
pipelines:
  setup_channel:
    steps:
      - name: create_chan
        type: step.teams_create_channel
        config:
          team_id: "team123"
          display_name: "announcements"
      - name: send_welcome
        type: step.teams_send_message
        config:
          team_id: "team123"
          channel_id: "newchan001"
          content: "Welcome to the channel!"
      - name: finalize
        type: step.set
        config:
          values:
            setup_complete: true
`),
		wftest.MockStep("step.teams_create_channel", wftest.Returns(map[string]any{
			"channel_id":   "newchan001",
			"id":           "newchan001",
			"display_name": "announcements",
			"team_id":      "team123",
			"web_url":      "https://teams.microsoft.com/channel/newchan001",
		})),
		wftest.MockStep("step.teams_send_message", wftest.Returns(map[string]any{
			"message_id": "welcomemsg",
			"id":         "welcomemsg",
			"team_id":    "team123",
			"channel_id": "newchan001",
		})),
	)

	result := h.ExecutePipeline("setup_channel", nil)
	if result.Error != nil {
		t.Fatal(result.Error)
	}
	if result.Output["setup_complete"] != true {
		t.Errorf("expected setup_complete=true, got %v", result.Output["setup_complete"])
	}

	chanOut := result.StepOutput("create_chan")
	if chanOut["channel_id"] != "newchan001" {
		t.Errorf("expected channel_id=newchan001, got %v", chanOut["channel_id"])
	}

	sendOut := result.StepOutput("send_welcome")
	if sendOut["message_id"] != "welcomemsg" {
		t.Errorf("expected message_id=welcomemsg, got %v", sendOut["message_id"])
	}
}
