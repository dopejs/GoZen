package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/dopejs/gozen/internal/bot/adapters"
)

// testSocketCounter generates unique short socket paths for macOS compatibility.
// Unix sockets on macOS have a 104-byte path limit.
var testSocketCounter atomic.Int64

func testSocketPath(t *testing.T) string {
	t.Helper()
	n := testSocketCounter.Add(1)
	path := fmt.Sprintf("/tmp/zen-test-%d.sock", n)
	t.Cleanup(func() { os.Remove(path) })
	return path
}

func TestGateway_StartStop(t *testing.T) {
	socketPath := testSocketPath(t)
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Error("socket file should exist after Start")
	}

	if err := g.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}

	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Error("socket file should be removed after Stop")
	}
}

func TestGateway_StartStop_WithConnections(t *testing.T) {
	socketPath := testSocketPath(t)
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	regPayload, _ := json.Marshal(RegisterPayload{
		ProcessID:   "test-proc-1",
		ProcessPath: "/path/to/test",
		PID:         12345,
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCRegister,
		Payload: regPayload,
	})

	time.Sleep(50 * time.Millisecond)

	if err := g.Stop(); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestGateway_handleConnection_Register(t *testing.T) {
	socketPath := testSocketPath(t)
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer g.Stop()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}

	regPayload, _ := json.Marshal(RegisterPayload{
		ProcessID:   "proc-reg-test",
		ProcessPath: "/path/to/project",
		PID:         99999,
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCRegister,
		Payload: regPayload,
	})

	time.Sleep(50 * time.Millisecond)

	info := g.registry.Find("proc-reg-test")
	if info == nil {
		t.Fatal("process should be registered")
	}
	if info.Path != "/path/to/project" {
		t.Errorf("expected path '/path/to/project', got '%s'", info.Path)
	}

	// Close connection - should trigger unregister
	conn.Close()
	time.Sleep(50 * time.Millisecond)

	info = g.registry.Find("proc-reg-test")
	if info != nil {
		t.Error("process should be unregistered after disconnect")
	}
}

func TestGateway_handleConnection_Heartbeat(t *testing.T) {
	socketPath := testSocketPath(t)
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer g.Stop()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	regPayload, _ := json.Marshal(RegisterPayload{
		ProcessID:   "proc-hb-test",
		ProcessPath: "/path/to/hb",
		PID:         11111,
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCRegister,
		Payload: regPayload,
	})
	time.Sleep(50 * time.Millisecond)

	hbPayload, _ := json.Marshal(HeartbeatPayload{
		ProcessID:   "proc-hb-test",
		Status:      "busy",
		CurrentTask: "running tests",
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCHeartbeat,
		Payload: hbPayload,
	})
	time.Sleep(50 * time.Millisecond)

	info := g.registry.Find("proc-hb-test")
	if info == nil {
		t.Fatal("process should be registered")
	}
	if info.Status != "busy" {
		t.Errorf("expected status 'busy', got '%s'", info.Status)
	}
	if info.CurrentTask != "running tests" {
		t.Errorf("expected task 'running tests', got '%s'", info.CurrentTask)
	}
}

func TestGateway_handleConnection_Notification(t *testing.T) {
	socketPath := testSocketPath(t)
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer g.Stop()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	regPayload, _ := json.Marshal(RegisterPayload{
		ProcessID:   "proc-notif-test",
		ProcessPath: "/path/to/notif",
		PID:         22222,
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCRegister,
		Payload: regPayload,
	})
	time.Sleep(50 * time.Millisecond)

	notifPayload, _ := json.Marshal(NotificationPayload{
		Level:   NotifyInfo,
		Title:   "Test",
		Message: "Test notification",
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCNotification,
		Payload: notifPayload,
	})
	time.Sleep(50 * time.Millisecond)
}

func TestGateway_handleConnection_Approval(t *testing.T) {
	socketPath := testSocketPath(t)
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer g.Stop()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	regPayload, _ := json.Marshal(RegisterPayload{
		ProcessID:   "proc-approval-test",
		ProcessPath: "/path/to/approval",
		PID:         33333,
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCRegister,
		Payload: regPayload,
	})
	time.Sleep(50 * time.Millisecond)

	approvalPayload, _ := json.Marshal(ApprovalPayload{
		ID:          "test-approval-1",
		Action:      "delete",
		Description: "Delete temp files",
		Timeout:     60,
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCApproval,
		Payload: approvalPayload,
	})
	time.Sleep(50 * time.Millisecond)
}

func TestGateway_handleConnection_Response(t *testing.T) {
	socketPath := testSocketPath(t)
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer g.Stop()

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	regPayload, _ := json.Marshal(RegisterPayload{
		ProcessID:   "proc-resp-test",
		ProcessPath: "/path/to/resp",
		PID:         44444,
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:    IPCRegister,
		Payload: regPayload,
	})
	time.Sleep(50 * time.Millisecond)

	respPayload, _ := json.Marshal(ResponsePayload{
		Success: true,
		Message: "Task completed",
	})
	json.NewEncoder(conn).Encode(IPCMessage{
		Type:      IPCResponse,
		RequestID: "req-123",
		Payload:   respPayload,
	})
	time.Sleep(50 * time.Millisecond)
}

func TestGateway_Start_InvalidSocketPath(t *testing.T) {
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: "/nonexistent/dir/test.sock",
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx := context.Background()
	err := g.Start(ctx)
	if err == nil {
		g.Stop()
		t.Error("Start should fail with invalid socket path")
	}
}

func TestGateway_Stop_NilFields(t *testing.T) {
	config := &GatewayConfig{
		Enabled: true,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	err := g.Stop()
	if err != nil {
		t.Errorf("Stop without Start should succeed: %v", err)
	}
}

func TestGateway_initAdapters_NoAdapters(t *testing.T) {
	config := &GatewayConfig{
		Enabled: true,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)
	g.ctx, g.cancel = context.WithCancel(context.Background())
	defer g.cancel()

	err := g.initAdapters()
	if err != nil {
		t.Errorf("initAdapters should succeed with no configs: %v", err)
	}
	if len(g.adapters) != 0 {
		t.Errorf("expected 0 adapters, got %d", len(g.adapters))
	}
}

func TestGateway_initAdapters_DisabledPlatforms(t *testing.T) {
	config := &GatewayConfig{
		Enabled: true,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
		Platforms: PlatformsConfig{
			Telegram:    &adapters.TelegramConfig{AdapterConfig: adapters.AdapterConfig{Enabled: false}, Token: "fake"},
			Discord:     &adapters.DiscordConfig{AdapterConfig: adapters.AdapterConfig{Enabled: false}, Token: "fake"},
			Slack:       &adapters.SlackConfig{AdapterConfig: adapters.AdapterConfig{Enabled: false}, BotToken: "fake"},
			Lark:        &adapters.LarkConfig{AdapterConfig: adapters.AdapterConfig{Enabled: false}, AppID: "fake"},
			FBMessenger: &adapters.FBMessengerConfig{AdapterConfig: adapters.AdapterConfig{Enabled: false}, PageToken: "fake"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)
	g.ctx, g.cancel = context.WithCancel(context.Background())
	defer g.cancel()

	err := g.initAdapters()
	if err != nil {
		t.Errorf("initAdapters should succeed: %v", err)
	}
	if len(g.adapters) != 0 {
		t.Errorf("expected 0 adapters (all disabled), got %d", len(g.adapters))
	}
}

func TestGateway_initAdapters_EmptyTokens(t *testing.T) {
	config := &GatewayConfig{
		Enabled: true,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
		Platforms: PlatformsConfig{
			Telegram:    &adapters.TelegramConfig{AdapterConfig: adapters.AdapterConfig{Enabled: true}, Token: ""},
			Discord:     &adapters.DiscordConfig{AdapterConfig: adapters.AdapterConfig{Enabled: true}, Token: ""},
			Slack:       &adapters.SlackConfig{AdapterConfig: adapters.AdapterConfig{Enabled: true}, BotToken: ""},
			Lark:        &adapters.LarkConfig{AdapterConfig: adapters.AdapterConfig{Enabled: true}, AppID: ""},
			FBMessenger: &adapters.FBMessengerConfig{AdapterConfig: adapters.AdapterConfig{Enabled: true}, PageToken: ""},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)
	g.ctx, g.cancel = context.WithCancel(context.Background())
	defer g.cancel()

	err := g.initAdapters()
	if err != nil {
		t.Errorf("initAdapters should succeed: %v", err)
	}
	if len(g.adapters) != 0 {
		t.Errorf("expected 0 adapters (empty tokens), got %d", len(g.adapters))
	}
}

func TestGateway_cleanupLoop_Cancel(t *testing.T) {
	config := &GatewayConfig{
		Enabled: true,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	g.ctx = ctx

	g.wg.Add(1)
	go g.cleanupLoop()

	cancel()
	g.wg.Wait()
}

func TestGateway_acceptConnections_Cancel(t *testing.T) {
	socketPath := testSocketPath(t)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}

	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	g.ctx = ctx
	g.listener = listener

	g.wg.Add(1)
	go g.acceptConnections()

	cancel()
	listener.Close()
	g.wg.Wait()
}

func TestGateway_handleConnection_ContextCancel(t *testing.T) {
	config := &GatewayConfig{
		Enabled: true,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	g.ctx = ctx

	server, client := net.Pipe()
	defer server.Close()

	g.wg.Add(1)
	go g.handleConnection(client)

	cancel()
	server.Close()
	g.wg.Wait()
}

func TestGateway_MultipleClientConnect(t *testing.T) {
	socketPath := testSocketPath(t)
	config := &GatewayConfig{
		Enabled:    true,
		SocketPath: socketPath,
		Interaction: InteractionConfig{
			MentionKeywords: []string{"@zen"},
		},
	}

	logger := log.New(os.Stderr, "[test] ", log.LstdFlags)
	g := NewGateway(config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer g.Stop()

	conn1, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("client 1 failed to connect: %v", err)
	}
	defer conn1.Close()

	conn2, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("client 2 failed to connect: %v", err)
	}
	defer conn2.Close()

	reg1, _ := json.Marshal(RegisterPayload{ProcessID: "proc-multi-1", ProcessPath: "/path/1", PID: 1})
	reg2, _ := json.Marshal(RegisterPayload{ProcessID: "proc-multi-2", ProcessPath: "/path/2", PID: 2})

	json.NewEncoder(conn1).Encode(IPCMessage{Type: IPCRegister, Payload: reg1})
	json.NewEncoder(conn2).Encode(IPCMessage{Type: IPCRegister, Payload: reg2})

	time.Sleep(100 * time.Millisecond)

	processes := g.registry.List()
	if len(processes) != 2 {
		t.Errorf("expected 2 registered processes, got %d", len(processes))
	}
}
