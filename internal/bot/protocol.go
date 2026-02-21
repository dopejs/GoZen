package bot

import (
	"encoding/json"

	"github.com/dopejs/gozen/internal/bot/adapters"
)

// Re-export types from adapters for convenience
type Platform = adapters.Platform

const (
	PlatformTelegram    = adapters.PlatformTelegram
	PlatformDiscord     = adapters.PlatformDiscord
	PlatformSlack       = adapters.PlatformSlack
	PlatformLark        = adapters.PlatformLark
	PlatformFBMessenger = adapters.PlatformFBMessenger
)

type Message = adapters.Message
type OutgoingMessage = adapters.OutgoingMessage
type Button = adapters.Button
type ButtonClick = adapters.ButtonClick

// ReplyContext contains info for replying to a message.
type ReplyContext struct {
	Platform  Platform `json:"platform"`
	ChatID    string   `json:"chat_id"`
	MessageID string   `json:"message_id,omitempty"`
	ThreadID  string   `json:"thread_id,omitempty"`
}

// --- IPC Protocol ---

// IPCMessageType defines IPC message types.
type IPCMessageType string

const (
	IPCRegister     IPCMessageType = "register"
	IPCUnregister   IPCMessageType = "unregister"
	IPCHeartbeat    IPCMessageType = "heartbeat"
	IPCCommand      IPCMessageType = "command"
	IPCResponse     IPCMessageType = "response"
	IPCNotification IPCMessageType = "notification"
	IPCApproval     IPCMessageType = "approval"
	IPCApprovalResp IPCMessageType = "approval_response"
)

// IPCMessage is the base IPC message structure.
type IPCMessage struct {
	Type      IPCMessageType  `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload"`
}

// RegisterPayload is sent when a process registers.
type RegisterPayload struct {
	ProcessID   string `json:"process_id"`
	ProcessPath string `json:"process_path"`
	SocketPath  string `json:"socket_path"`
	PID         int    `json:"pid"`
}

// HeartbeatPayload is sent periodically by processes.
type HeartbeatPayload struct {
	ProcessID   string `json:"process_id"`
	Status      string `json:"status"`
	CurrentTask string `json:"current_task,omitempty"`
	Memory      uint64 `json:"memory,omitempty"`
}

// CommandPayload is sent from gateway to process.
type CommandPayload struct {
	Intent  *ParsedIntent `json:"intent"`
	User    UserInfo      `json:"user"`
	ReplyTo ReplyContext  `json:"reply_to"`
}

// UserInfo contains info about the message sender.
type UserInfo struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Platform Platform `json:"platform"`
}

// ResponsePayload is sent from process to gateway.
type ResponsePayload struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Format  string      `json:"format,omitempty"` // plain, markdown
	Data    interface{} `json:"data,omitempty"`
	Buttons []Button    `json:"buttons,omitempty"`
}

// NotificationPayload is sent from process to gateway.
type NotificationPayload struct {
	Level   string   `json:"level"` // info, warning, error, success
	Title   string   `json:"title"`
	Message string   `json:"message"`
	Buttons []Button `json:"buttons,omitempty"`
}

// ApprovalPayload is sent from process to gateway.
type ApprovalPayload struct {
	ID          string `json:"id"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Details     string `json:"details,omitempty"`
	Timeout     int    `json:"timeout,omitempty"` // seconds, 0 = no timeout
}

// ApprovalResponsePayload is sent from gateway to process.
type ApprovalResponsePayload struct {
	RequestID string `json:"request_id"`
	Approved  bool   `json:"approved"`
	Comment   string `json:"comment,omitempty"`
	UserID    string `json:"user_id"`
}

// --- Intent ---

// Intent represents a parsed user intent.
type Intent string

const (
	IntentQueryStatus Intent = "query_status"
	IntentQueryList   Intent = "query_list"
	IntentSendTask    Intent = "send_task"
	IntentControl     Intent = "control"
	IntentApprove     Intent = "approve"
	IntentSubscribe   Intent = "subscribe"
	IntentBind        Intent = "bind"
	IntentHelp        Intent = "help"
	IntentUnknown     Intent = "unknown"
)

// ParsedIntent represents the result of intent parsing.
type ParsedIntent struct {
	Intent   Intent            `json:"intent"`
	Target   string            `json:"target,omitempty"`
	Action   string            `json:"action,omitempty"`
	Task     string            `json:"task,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
	Raw      string            `json:"raw"`
	Approved *bool             `json:"approved,omitempty"` // for approval responses
}

// --- Notification Levels ---

const (
	NotifyInfo    = "info"
	NotifyWarning = "warning"
	NotifyError   = "error"
	NotifySuccess = "success"
)
