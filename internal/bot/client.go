package bot

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Client is the bot client for zen processes to communicate with the gateway.
type Client struct {
	processID   string
	processPath string
	socketPath  string
	gatewayPath string
	conn        net.Conn
	mu          sync.Mutex
	handlers    ClientHandlers
	connected   bool
}

// ClientHandlers contains handlers for incoming messages from gateway.
type ClientHandlers struct {
	OnCommand  func(*CommandPayload) *ResponsePayload
	OnApproval func(*ApprovalResponsePayload)
}

// NewClient creates a new bot client.
func NewClient(processPath, gatewayPath string) *Client {
	if gatewayPath == "" {
		gatewayPath = filepath.Join(os.TempDir(), "zen-gateway.sock")
	}

	processID := fmt.Sprintf("zen-%d", os.Getpid())

	return &Client{
		processID:   processID,
		processPath: processPath,
		gatewayPath: gatewayPath,
	}
}

// SetHandlers sets the handlers for incoming messages.
func (c *Client) SetHandlers(handlers ClientHandlers) {
	c.handlers = handlers
}

// Connect connects to the gateway and registers this process.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	// Check if gateway socket exists
	if _, err := os.Stat(c.gatewayPath); os.IsNotExist(err) {
		return fmt.Errorf("gateway not running (socket not found: %s)", c.gatewayPath)
	}

	conn, err := net.Dial("unix", c.gatewayPath)
	if err != nil {
		return fmt.Errorf("failed to connect to gateway: %w", err)
	}
	c.conn = conn

	// Send registration
	payload := RegisterPayload{
		ProcessID:   c.processID,
		ProcessPath: c.processPath,
		PID:         os.Getpid(),
	}

	if err := c.sendMessage(IPCRegister, "", payload); err != nil {
		conn.Close()
		return fmt.Errorf("failed to register: %w", err)
	}

	c.connected = true

	// Start receiving messages
	go c.receiveLoop()

	// Start heartbeat
	go c.heartbeatLoop()

	return nil
}

// Disconnect disconnects from the gateway.
func (c *Client) Disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return
	}

	c.sendMessage(IPCUnregister, "", nil)
	c.conn.Close()
	c.connected = false
}

// IsConnected returns whether the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// UpdateStatus updates the process status.
func (c *Client) UpdateStatus(status, currentTask string) error {
	payload := HeartbeatPayload{
		ProcessID:   c.processID,
		Status:      status,
		CurrentTask: currentTask,
	}
	return c.sendMessage(IPCHeartbeat, "", payload)
}

// SendNotification sends a notification to the gateway.
func (c *Client) SendNotification(level, title, message string) error {
	payload := NotificationPayload{
		Level:   level,
		Title:   title,
		Message: message,
	}
	return c.sendMessage(IPCNotification, "", payload)
}

// SendNotificationWithButtons sends a notification with action buttons.
func (c *Client) SendNotificationWithButtons(level, title, message string, buttons []Button) error {
	payload := NotificationPayload{
		Level:   level,
		Title:   title,
		Message: message,
		Buttons: buttons,
	}
	return c.sendMessage(IPCNotification, "", payload)
}

// RequestApproval sends an approval request and waits for response.
func (c *Client) RequestApproval(id, action, description, details string, timeout int) error {
	payload := ApprovalPayload{
		ID:          id,
		Action:      action,
		Description: description,
		Details:     details,
		Timeout:     timeout,
	}
	return c.sendMessage(IPCApproval, id, payload)
}

// SendResponse sends a response to a command.
func (c *Client) SendResponse(requestID string, success bool, message string) error {
	payload := ResponsePayload{
		Success: success,
		Message: message,
	}
	return c.sendMessage(IPCResponse, requestID, payload)
}

func (c *Client) sendMessage(msgType IPCMessageType, requestID string, payload interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.conn == nil {
		return fmt.Errorf("not connected")
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := IPCMessage{
		Type:      msgType,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	return json.NewEncoder(c.conn).Encode(msg)
}

func (c *Client) receiveLoop() {
	decoder := json.NewDecoder(c.conn)

	for {
		var msg IPCMessage
		if err := decoder.Decode(&msg); err != nil {
			c.mu.Lock()
			c.connected = false
			c.mu.Unlock()
			return
		}

		switch msg.Type {
		case IPCCommand:
			if c.handlers.OnCommand != nil {
				var payload CommandPayload
				json.Unmarshal(msg.Payload, &payload)
				response := c.handlers.OnCommand(&payload)
				if response != nil {
					c.sendMessage(IPCResponse, msg.RequestID, response)
				}
			}

		case IPCApprovalResp:
			if c.handlers.OnApproval != nil {
				var payload ApprovalResponsePayload
				json.Unmarshal(msg.Payload, &payload)
				c.handlers.OnApproval(&payload)
			}
		}
	}
}

func (c *Client) heartbeatLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		connected := c.connected
		c.mu.Unlock()

		if !connected {
			return
		}

		c.sendMessage(IPCHeartbeat, "", HeartbeatPayload{
			ProcessID: c.processID,
			Status:    "idle",
		})
	}
}
