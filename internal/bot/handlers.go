package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dopejs/gozen/internal/bot/adapters"
)

// handleMessage processes incoming messages from any platform.
func (g *Gateway) handleMessage(msg *Message) {
	// Determine if we should respond
	requireMention := g.config.Interaction.RequireMention
	if msg.IsDirectMsg {
		// Direct messages: check direct_message_mode
		if g.config.Interaction.DirectMsgMode == "always" {
			requireMention = false
		}
	} else {
		// Channel messages: check channel_mode
		if g.config.Interaction.ChannelMode == "always" {
			requireMention = false
		}
	}

	// Parse intent
	intent := g.nlu.Parse(msg, requireMention)
	if intent == nil {
		return
	}

	// Get or create session
	session := g.sessions.GetOrCreate(msg.Platform, msg.UserID, msg.ChatID)

	// Build reply context
	replyTo := ReplyContext{
		Platform:  msg.Platform,
		ChatID:    msg.ChatID,
		MessageID: msg.ID,
		ThreadID:  msg.ThreadID,
	}

	// Handle the intent
	g.processIntent(intent, session, replyTo, msg)
}

// extractMissingParams uses LLM to extract parameters for intents that need them.
func (g *Gateway) extractMissingParams(intent *ParsedIntent, msg *Message) (*ParsedIntent, error) {
	// Check if we have a classifier
	if g.skillMatcher == nil || g.skillMatcher.Classifier() == nil {
		return intent, nil // no LLM available, keep as-is
	}

	// Determine if parameters are missing based on intent type
	switch intent.Intent {
	case IntentControl:
		if intent.Action == "" {
			// Extract action and target from message
			params, err := g.skillMatcher.Classifier().ExtractParams(context.Background(), msg.Content, intent.Intent)
			if err != nil {
				return intent, fmt.Errorf("extract control params: %w", err)
			}
			if action, ok := params["action"]; ok {
				intent.Action = action
			}
			if target, ok := params["target"]; ok {
				intent.Target = target
			}
		}
	case IntentBind:
		if intent.Target == "" {
			params, err := g.skillMatcher.Classifier().ExtractParams(context.Background(), msg.Content, intent.Intent)
			if err != nil {
				return intent, fmt.Errorf("extract bind params: %w", err)
			}
			if target, ok := params["target"]; ok {
				intent.Target = target
			}
		}
	case IntentSendTask:
		if intent.Target == "" || intent.Task == "" {
			params, err := g.skillMatcher.Classifier().ExtractParams(context.Background(), msg.Content, intent.Intent)
			if err != nil {
				return intent, fmt.Errorf("extract send_task params: %w", err)
			}
			if target, ok := params["target"]; ok {
				intent.Target = target
			}
			if task, ok := params["task"]; ok {
				intent.Task = task
			}
		}
	// Other intents may not need parameter extraction
	}
	return intent, nil
}
// processIntent handles a parsed intent.
func (g *Gateway) processIntent(intent *ParsedIntent, session *Session, replyTo ReplyContext, msg *Message) {
	// Try to extract missing parameters via LLM if available
	updatedIntent, err := g.extractMissingParams(intent, msg)
	if err != nil {
		g.logger.Printf("Parameter extraction failed: %v", err)
		// Continue with original intent
		updatedIntent = intent
	}
	intent = updatedIntent

	switch intent.Intent {
	case IntentControl:
		g.handleControl(intent, session, replyTo)

	case IntentBind:
		g.handleBind(intent, session, replyTo)

	case IntentApprove:
		g.handleApprovalResponse(intent, session, replyTo, msg)

	case IntentSendTask:
		g.handleSendTask(intent, session, replyTo)

	case IntentPersona:
		g.handlePersona(intent, replyTo)

	case IntentForget:
		g.handleForget(session, replyTo)

	default:
		// All other intents (including Help, Chat, QueryStatus, Unknown)
		// go through the LLM with full process context
		g.handleChat(intent, session, replyTo)
	}
}

// handleChat handles chat intent using LLM with conversation history.
func (g *Gateway) handleChat(intent *ParsedIntent, session *Session, replyTo ReplyContext) {
	// If no LLM client configured, send a friendly fallback
	if g.llm == nil {
		g.sendMessage(replyTo, &OutgoingMessage{
			Text: "Hi! I'm Zen, the GoZen assistant. I can help you with:\n\n" +
				"• `bind <name>` - Bind to a process\n" +
				"• `pause/resume/cancel [name]` - Control tasks\n" +
				"• `send <name> <task>` or `<name>: <task>` - Send a task\n" +
				"• `persona <text>` - Set bot persona\n" +
				"• `forget` - Clear conversation history\n\n" +
				"What would you like to do?",
			Format: "markdown",
		})
		return
	}

	// Build system prompt with full process state
	processes := g.ListAllProcesses()
	memory, _ := LoadMemory(g.config.MemoryDir)
	systemPrompt := BuildSystemPrompt(processes, g.config.Profile, memory)

	// Get user message
	userMessage := intent.Raw
	if intent.Task != "" {
		userMessage = intent.Task
	}
	if userMessage == "" {
		userMessage = "Hello"
	}

	// Add user message to history
	if session.History != nil {
		session.History.Add(ChatMessage{Role: "user", Content: userMessage})
	}

	// Get conversation history
	var history []ChatMessage
	if session.History != nil {
		history = session.History.Messages()
	} else {
		history = []ChatMessage{{Role: "user", Content: userMessage}}
	}

	// Call LLM with full history
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	response, err := g.llm.Chat(ctx, systemPrompt, history)
	if err != nil {
		g.logger.Printf("LLM chat error: %v", err)
		// Fallback to simple response
		g.sendMessage(replyTo, &OutgoingMessage{
			Text: "Hi! I'm Zen. How can I help you today? Try asking about connected processes or their status.",
		})
		return
	}

	// Add assistant response to history
	if session.History != nil {
		session.History.Add(ChatMessage{Role: "assistant", Content: response})
	}

	g.sendMessage(replyTo, &OutgoingMessage{
		Text:   response,
		Format: "markdown",
	})
}

// handleControl handles control commands.
func (g *Gateway) handleControl(intent *ParsedIntent, session *Session, replyTo ReplyContext) {
	target := intent.Target
	if target == "" {
		target = session.BoundProcess
	}

	if target == "" {
		processes := g.registry.List()
		if len(processes) == 1 {
			target = processes[0].Name
		} else {
			g.sendMessage(replyTo, &OutgoingMessage{
				Text: "Please specify which process or use `bind <name>` first.",
			})
			return
		}
	}

	process := g.registry.Find(target)
	if process == nil {
		g.sendMessage(replyTo, &OutgoingMessage{
			Text: fmt.Sprintf("Process `%s` not found.", target),
		})
		return
	}

	// Send control command to process
	g.sendCommandToProcess(process.ID, intent, replyTo)
}

// handleBind handles bind command.
func (g *Gateway) handleBind(intent *ParsedIntent, session *Session, replyTo ReplyContext) {
	if intent.Target == "" {
		// Show current binding
		if session.BoundProcess != "" {
			g.sendMessage(replyTo, &OutgoingMessage{
				Text: fmt.Sprintf("Currently bound to `%s`.", session.BoundProcess),
			})
		} else {
			g.sendMessage(replyTo, &OutgoingMessage{
				Text: "Not bound to any process. Use `bind <name>` to bind.",
			})
		}
		return
	}

	process := g.registry.Find(intent.Target)
	if process == nil {
		g.sendMessage(replyTo, &OutgoingMessage{
			Text: fmt.Sprintf("Process `%s` not found.", intent.Target),
		})
		return
	}

	g.sessions.Bind(replyTo.Platform, session.UserID, process.Name)
	g.sendMessage(replyTo, &OutgoingMessage{
		Text: fmt.Sprintf("Bound to `%s`. Subsequent commands will target this process.", process.Name),
	})
}

// handleSendTask sends a task to a process.
func (g *Gateway) handleSendTask(intent *ParsedIntent, session *Session, replyTo ReplyContext) {
	target := intent.Target
	if target == "" {
		target = session.BoundProcess
	}

	if target == "" {
		processes := g.registry.List()
		if len(processes) == 1 {
			target = processes[0].Name
		} else if len(processes) == 0 {
			// No processes - fall back to chat
			g.handleChat(intent, session, replyTo)
			return
		} else {
			g.sendMessage(replyTo, &OutgoingMessage{
				Text: "Multiple processes available. Please specify which one (e.g., `send api run tests` or `api: run tests`).",
			})
			return
		}
	}

	process := g.registry.Find(target)
	if process == nil {
		g.sendMessage(replyTo, &OutgoingMessage{
			Text: fmt.Sprintf("Process `%s` not found.", target),
		})
		return
	}

	g.sendCommandToProcess(process.ID, intent, replyTo)
	g.sendMessage(replyTo, &OutgoingMessage{
		Text: fmt.Sprintf("Task sent to `%s`.", process.Name),
	})
}

// handlePersona handles persona set/show/clear commands.
func (g *Gateway) handlePersona(intent *ParsedIntent, replyTo ReplyContext) {
	memDir := g.config.MemoryDir

	switch intent.Action {
	case "show":
		content, err := LoadMemory(memDir)
		if err != nil {
			g.sendMessage(replyTo, &OutgoingMessage{Text: fmt.Sprintf("Failed to read memory: %v", err)})
			return
		}
		if content == "" {
			g.sendMessage(replyTo, &OutgoingMessage{Text: "No persona set. Use `persona <text>` to set one."})
			return
		}
		g.sendMessage(replyTo, &OutgoingMessage{
			Text:   fmt.Sprintf("**Current Persona**\n\n%s", content),
			Format: "markdown",
		})

	case "set":
		if err := SaveMemory(memDir, intent.Task); err != nil {
			g.sendMessage(replyTo, &OutgoingMessage{Text: fmt.Sprintf("Failed to save persona: %v", err)})
			return
		}
		g.sendMessage(replyTo, &OutgoingMessage{Text: "Persona updated."})

	case "clear":
		if err := ClearMemory(memDir); err != nil {
			g.sendMessage(replyTo, &OutgoingMessage{Text: fmt.Sprintf("Failed to clear persona: %v", err)})
			return
		}
		g.sendMessage(replyTo, &OutgoingMessage{Text: "Persona cleared."})
	}
}

// handleForget clears the session's conversation history.
func (g *Gateway) handleForget(session *Session, replyTo ReplyContext) {
	if session.History != nil {
		session.History.Clear()
	}
	g.sendMessage(replyTo, &OutgoingMessage{Text: "Conversation history cleared."})
}

// handleApprovalResponse handles approval/rejection responses.
func (g *Gateway) handleApprovalResponse(intent *ParsedIntent, session *Session, replyTo ReplyContext, msg *Message) {
	// Find pending approval by reply context
	var approval *PendingApproval

	// Check if replying to an approval message
	if msg.ReplyTo != "" {
		approval = g.approvals.GetByMessageID(msg.ReplyTo)
	}

	if approval == nil {
		g.sendMessage(replyTo, &OutgoingMessage{
			Text: "No pending approval found. Please reply to an approval request or click the buttons.",
		})
		return
	}

	// Send response to process
	response := ApprovalResponsePayload{
		RequestID: approval.ID,
		Approved:  intent.Approved != nil && *intent.Approved,
		UserID:    session.UserID,
	}

	g.sendIPCMessage(approval.ProcessID, IPCApprovalResp, approval.ID, response)
	g.approvals.Remove(approval.ID)

	status := "rejected"
	if response.Approved {
		status = "approved"
	}
	g.sendMessage(replyTo, &OutgoingMessage{
		Text: fmt.Sprintf("Request %s.", status),
	})
}

// handleButtonClick processes button click events.
func (g *Gateway) handleButtonClick(click *ButtonClick) {
	// Check if it's an approval button
	if strings.HasPrefix(click.ButtonID, "approve_") || strings.HasPrefix(click.ButtonID, "reject_") {
		approvalID := click.Data
		approval := g.approvals.Get(approvalID)
		if approval == nil {
			return
		}

		approved := strings.HasPrefix(click.ButtonID, "approve_")
		response := ApprovalResponsePayload{
			RequestID: approvalID,
			Approved:  approved,
			UserID:    click.UserID,
		}

		g.sendIPCMessage(approval.ProcessID, IPCApprovalResp, approvalID, response)
		g.approvals.Remove(approvalID)

		// Update the message to show result
		status := "❌ Rejected"
		if approved {
			status = "✅ Approved"
		}

		replyTo := approval.ReplyTo
		g.editMessage(replyTo, approval.MessageID, &OutgoingMessage{
			Text:   fmt.Sprintf("%s\n\n%s by <@%s>", status, status, click.UserID),
			Format: "markdown",
		})
	}
}

// handleNotification handles notifications from processes.
func (g *Gateway) handleNotification(processID string, payload *NotificationPayload) {
	process := g.registry.Get(processID)
	if process == nil {
		return
	}

	// Check quiet hours
	if g.isQuietHours() && payload.Level != NotifyError {
		return
	}

	// Build notification message
	icon := "ℹ️"
	switch payload.Level {
	case NotifyWarning:
		icon = "⚠️"
	case NotifyError:
		icon = "🔴"
	case NotifySuccess:
		icon = "✅"
	}

	text := fmt.Sprintf("%s **%s** [%s]\n\n%s", icon, payload.Title, process.Name, payload.Message)

	// Send to default chat if configured
	if g.config.Notifications.DefaultChat != nil {
		replyTo := ReplyContext{
			Platform: g.config.Notifications.DefaultChat.Platform,
			ChatID:   g.config.Notifications.DefaultChat.ChatID,
		}
		g.sendMessage(replyTo, &OutgoingMessage{
			Text:    text,
			Format:  "markdown",
			Buttons: payload.Buttons,
		})
	}
}

// handleApprovalRequest handles approval requests from processes.
func (g *Gateway) handleApprovalRequest(processID string, payload *ApprovalPayload) {
	process := g.registry.Get(processID)
	if process == nil {
		return
	}

	if g.config.Notifications.DefaultChat == nil {
		g.logger.Println("Cannot send approval request: no default chat configured")
		return
	}

	replyTo := ReplyContext{
		Platform: g.config.Notifications.DefaultChat.Platform,
		ChatID:   g.config.Notifications.DefaultChat.ChatID,
	}

	text := fmt.Sprintf("🔔 **Approval Request** [%s]\n\n**Action:** %s\n\n%s",
		process.Name, payload.Action, payload.Description)

	if payload.Details != "" {
		text += fmt.Sprintf("\n\n```\n%s\n```", payload.Details)
	}

	buttons := []Button{
		{ID: "approve_" + payload.ID, Label: "✅ Approve", Style: "primary", Data: payload.ID},
		{ID: "reject_" + payload.ID, Label: "❌ Reject", Style: "danger", Data: payload.ID},
	}

	msgID, _ := g.sendMessage(replyTo, &OutgoingMessage{
		Text:    text,
		Format:  "markdown",
		Buttons: buttons,
	})

	// Track pending approval
	timeout := time.Time{}
	if payload.Timeout > 0 {
		timeout = time.Now().Add(time.Duration(payload.Timeout) * time.Second)
	}

	g.approvals.Add(&PendingApproval{
		ID:        payload.ID,
		ProcessID: processID,
		ReplyTo:   replyTo,
		MessageID: msgID,
		CreatedAt: time.Now(),
		Timeout:   timeout,
	})
}

// handleProcessResponse handles responses from processes.
func (g *Gateway) handleProcessResponse(requestID string, payload *ResponsePayload) {
	// TODO: implement request tracking for async responses
	_ = requestID
	_ = payload
}

// sendCommandToProcess sends a command to a process via IPC.
func (g *Gateway) sendCommandToProcess(processID string, intent *ParsedIntent, replyTo ReplyContext) {
	payload := CommandPayload{
		Intent:  intent,
		ReplyTo: replyTo,
	}
	g.sendIPCMessage(processID, IPCCommand, "", payload)
}

// sendIPCMessage sends an IPC message to a process.
func (g *Gateway) sendIPCMessage(processID string, msgType IPCMessageType, requestID string, payload interface{}) error {
	g.mu.RLock()
	conn, ok := g.connections[processID]
	g.mu.RUnlock()

	if !ok {
		return fmt.Errorf("process not connected: %s", processID)
	}

	payloadBytes, _ := json.Marshal(payload)
	msg := IPCMessage{
		Type:      msgType,
		RequestID: requestID,
		Payload:   payloadBytes,
	}

	return json.NewEncoder(conn).Encode(msg)
}

// sendMessage sends a message via the appropriate adapter.
func (g *Gateway) sendMessage(replyTo ReplyContext, msg *OutgoingMessage) (string, error) {
	adapter := g.getAdapter(replyTo.Platform)
	if adapter == nil {
		return "", fmt.Errorf("no adapter for platform: %s", replyTo.Platform)
	}

	if replyTo.MessageID != "" {
		return adapter.SendReply(replyTo.ChatID, replyTo.MessageID, msg)
	}
	return adapter.SendMessage(replyTo.ChatID, msg)
}

// editMessage edits a message via the appropriate adapter.
func (g *Gateway) editMessage(replyTo ReplyContext, msgID string, msg *OutgoingMessage) error {
	adapter := g.getAdapter(replyTo.Platform)
	if adapter == nil {
		return fmt.Errorf("no adapter for platform: %s", replyTo.Platform)
	}
	return adapter.EditMessage(replyTo.ChatID, msgID, msg)
}

// getAdapter returns the adapter for a platform.
func (g *Gateway) getAdapter(platform Platform) adapters.Adapter {
	for _, a := range g.adapters {
		if a.Platform() == adapters.Platform(platform) {
			return a
		}
	}
	return nil
}

// isQuietHours checks if current time is within quiet hours.
func (g *Gateway) isQuietHours() bool {
	qh := g.config.Notifications.QuietHours
	if qh == nil || !qh.Enabled {
		return false
	}

	loc, err := time.LoadLocation(qh.Timezone)
	if err != nil {
		loc = time.Local
	}

	now := time.Now().In(loc)
	nowMinutes := now.Hour()*60 + now.Minute()

	startParts := strings.Split(qh.Start, ":")
	endParts := strings.Split(qh.End, ":")

	if len(startParts) != 2 || len(endParts) != 2 {
		return false
	}

	var startH, startM, endH, endM int
	fmt.Sscanf(startParts[0], "%d", &startH)
	fmt.Sscanf(startParts[1], "%d", &startM)
	fmt.Sscanf(endParts[0], "%d", &endH)
	fmt.Sscanf(endParts[1], "%d", &endM)

	startMinutes := startH*60 + startM
	endMinutes := endH*60 + endM

	if startMinutes < endMinutes {
		return nowMinutes >= startMinutes && nowMinutes < endMinutes
	}
	// Crosses midnight
	return nowMinutes >= startMinutes || nowMinutes < endMinutes
}
