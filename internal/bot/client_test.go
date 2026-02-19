package bot

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("/path/to/project", "")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.processPath != "/path/to/project" {
		t.Errorf("expected processPath '/path/to/project', got '%s'", client.processPath)
	}
	if client.gatewayPath != filepath.Join(os.TempDir(), "zen-gateway.sock") {
		t.Errorf("unexpected default gatewayPath: %s", client.gatewayPath)
	}
}

func TestNewClient_CustomGatewayPath(t *testing.T) {
	client := NewClient("/path/to/project", "/custom/gateway.sock")
	if client.gatewayPath != "/custom/gateway.sock" {
		t.Errorf("expected gatewayPath '/custom/gateway.sock', got '%s'", client.gatewayPath)
	}
}

func TestClient_SetHandlers(t *testing.T) {
	client := NewClient("/path/to/project", "")

	handlers := ClientHandlers{
		OnCommand: func(cmd *CommandPayload) *ResponsePayload {
			return nil
		},
	}

	client.SetHandlers(handlers)

	if client.handlers.OnCommand == nil {
		t.Error("OnCommand handler should be set")
	}
}

func TestClient_IsConnected_Initial(t *testing.T) {
	client := NewClient("/path/to/project", "")
	if client.IsConnected() {
		t.Error("new client should not be connected")
	}
}

func TestClient_Connect_NoGateway(t *testing.T) {
	client := NewClient("/path/to/project", "/nonexistent/gateway.sock")
	err := client.Connect()
	if err == nil {
		t.Error("Connect should fail when gateway doesn't exist")
	}
}

func TestClient_Disconnect_NotConnected(t *testing.T) {
	client := NewClient("/path/to/project", "")
	// Should not panic
	client.Disconnect()
}

func TestClient_UpdateStatus_NotConnected(t *testing.T) {
	client := NewClient("/path/to/project", "")
	err := client.UpdateStatus("busy", "running tests")
	if err == nil {
		t.Error("UpdateStatus should fail when not connected")
	}
}

func TestClient_SendNotification_NotConnected(t *testing.T) {
	client := NewClient("/path/to/project", "")
	err := client.SendNotification(NotifyInfo, "Test", "Message")
	if err == nil {
		t.Error("SendNotification should fail when not connected")
	}
}

func TestClient_SendNotificationWithButtons_NotConnected(t *testing.T) {
	client := NewClient("/path/to/project", "")
	buttons := []Button{{ID: "btn1", Label: "OK"}}
	err := client.SendNotificationWithButtons(NotifyInfo, "Test", "Message", buttons)
	if err == nil {
		t.Error("SendNotificationWithButtons should fail when not connected")
	}
}

func TestClient_RequestApproval_NotConnected(t *testing.T) {
	client := NewClient("/path/to/project", "")
	err := client.RequestApproval("id", "action", "desc", "details", 300)
	if err == nil {
		t.Error("RequestApproval should fail when not connected")
	}
}

func TestClient_SendResponse_NotConnected(t *testing.T) {
	client := NewClient("/path/to/project", "")
	err := client.SendResponse("req-1", true, "done")
	if err == nil {
		t.Error("SendResponse should fail when not connected")
	}
}
