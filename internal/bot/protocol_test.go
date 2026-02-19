package bot

import (
	"encoding/json"
	"testing"
)

func TestIPCMessageTypes(t *testing.T) {
	// Test that all message types are defined
	types := []IPCMessageType{
		IPCRegister,
		IPCUnregister,
		IPCHeartbeat,
		IPCCommand,
		IPCResponse,
		IPCNotification,
		IPCApproval,
		IPCApprovalResp,
	}

	for _, typ := range types {
		if typ == "" {
			t.Error("IPCMessageType should not be empty")
		}
	}
}

func TestIntentTypes(t *testing.T) {
	// Test that all intent types are defined
	intents := []Intent{
		IntentQueryStatus,
		IntentQueryList,
		IntentSendTask,
		IntentControl,
		IntentApprove,
		IntentSubscribe,
		IntentBind,
		IntentHelp,
		IntentUnknown,
	}

	for _, intent := range intents {
		if intent == "" {
			t.Error("Intent should not be empty")
		}
	}
}

func TestRegisterPayload_JSON(t *testing.T) {
	payload := RegisterPayload{
		ProcessID:   "test-1",
		ProcessPath: "/path/to/project",
		SocketPath:  "/tmp/test.sock",
		PID:         1234,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded RegisterPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ProcessID != payload.ProcessID {
		t.Errorf("ProcessID mismatch: got %s, want %s", decoded.ProcessID, payload.ProcessID)
	}
	if decoded.PID != payload.PID {
		t.Errorf("PID mismatch: got %d, want %d", decoded.PID, payload.PID)
	}
}

func TestHeartbeatPayload_JSON(t *testing.T) {
	payload := HeartbeatPayload{
		ProcessID:   "test-1",
		Status:      "busy",
		CurrentTask: "running tests",
		Memory:      1024000,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded HeartbeatPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Status != payload.Status {
		t.Errorf("Status mismatch: got %s, want %s", decoded.Status, payload.Status)
	}
	if decoded.CurrentTask != payload.CurrentTask {
		t.Errorf("CurrentTask mismatch: got %s, want %s", decoded.CurrentTask, payload.CurrentTask)
	}
}

func TestCommandPayload_JSON(t *testing.T) {
	intent := &ParsedIntent{
		Intent: IntentSendTask,
		Target: "api",
		Task:   "run tests",
	}
	payload := CommandPayload{
		Intent: intent,
		User: UserInfo{
			ID:       "user1",
			Name:     "John",
			Platform: PlatformTelegram,
		},
		ReplyTo: ReplyContext{
			Platform: PlatformTelegram,
			ChatID:   "chat1",
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded CommandPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Intent.Target != "api" {
		t.Errorf("Intent.Target mismatch: got %s, want api", decoded.Intent.Target)
	}
	if decoded.User.Name != "John" {
		t.Errorf("User.Name mismatch: got %s, want John", decoded.User.Name)
	}
}

func TestResponsePayload_JSON(t *testing.T) {
	payload := ResponsePayload{
		Success: true,
		Message: "Task completed",
		Format:  "markdown",
		Buttons: []Button{
			{ID: "btn1", Label: "OK", Style: "primary"},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ResponsePayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !decoded.Success {
		t.Error("Success should be true")
	}
	if len(decoded.Buttons) != 1 {
		t.Errorf("Expected 1 button, got %d", len(decoded.Buttons))
	}
}

func TestNotificationPayload_JSON(t *testing.T) {
	payload := NotificationPayload{
		Level:   NotifyWarning,
		Title:   "Warning",
		Message: "Something happened",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded NotificationPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Level != NotifyWarning {
		t.Errorf("Level mismatch: got %s, want %s", decoded.Level, NotifyWarning)
	}
}

func TestApprovalPayload_JSON(t *testing.T) {
	payload := ApprovalPayload{
		ID:          "approval-1",
		Action:      "delete file",
		Description: "Delete important.txt",
		Details:     "rm -rf important.txt",
		Timeout:     300,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ApprovalPayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Timeout != 300 {
		t.Errorf("Timeout mismatch: got %d, want 300", decoded.Timeout)
	}
}

func TestApprovalResponsePayload_JSON(t *testing.T) {
	payload := ApprovalResponsePayload{
		RequestID: "approval-1",
		Approved:  true,
		Comment:   "Looks good",
		UserID:    "user1",
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ApprovalResponsePayload
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !decoded.Approved {
		t.Error("Approved should be true")
	}
}

func TestParsedIntent_JSON(t *testing.T) {
	approved := true
	intent := ParsedIntent{
		Intent:   IntentApprove,
		Target:   "api",
		Action:   "approve",
		Task:     "",
		Params:   map[string]string{"key": "value"},
		Raw:      "approve api",
		Approved: &approved,
	}

	data, err := json.Marshal(intent)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ParsedIntent
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Intent != IntentApprove {
		t.Errorf("Intent mismatch: got %s, want %s", decoded.Intent, IntentApprove)
	}
	if decoded.Approved == nil || !*decoded.Approved {
		t.Error("Approved should be true")
	}
}

func TestIPCMessage_JSON(t *testing.T) {
	payload := RegisterPayload{
		ProcessID: "test-1",
		PID:       1234,
	}
	payloadBytes, _ := json.Marshal(payload)

	msg := IPCMessage{
		Type:      IPCRegister,
		RequestID: "req-1",
		Payload:   payloadBytes,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded IPCMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Type != IPCRegister {
		t.Errorf("Type mismatch: got %s, want %s", decoded.Type, IPCRegister)
	}

	var decodedPayload RegisterPayload
	if err := json.Unmarshal(decoded.Payload, &decodedPayload); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}
	if decodedPayload.ProcessID != "test-1" {
		t.Errorf("Payload ProcessID mismatch: got %s, want test-1", decodedPayload.ProcessID)
	}
}

func TestReplyContext_JSON(t *testing.T) {
	ctx := ReplyContext{
		Platform:  PlatformDiscord,
		ChatID:    "channel-123",
		MessageID: "msg-456",
		ThreadID:  "thread-789",
	}

	data, err := json.Marshal(ctx)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded ReplyContext
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Platform != PlatformDiscord {
		t.Errorf("Platform mismatch: got %s, want %s", decoded.Platform, PlatformDiscord)
	}
	if decoded.ThreadID != "thread-789" {
		t.Errorf("ThreadID mismatch: got %s, want thread-789", decoded.ThreadID)
	}
}
