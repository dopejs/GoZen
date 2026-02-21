package bot

import (
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

// processIntent handles a parsed intent.
func (g *Gateway) processIntent(intent *ParsedIntent, session *Session, replyTo ReplyContext, msg *Message) {
	switch intent.Intent {
	case IntentHelp:
		g.sendHelp(replyTo)

	case IntentQueryList:
		g.sendProcessList(replyTo)

	case IntentQueryStatus:
		g.handleStatusQuery(intent, session, replyTo)

	case IntentControl:
		g.handleControl(intent, session, replyTo)

	case IntentBind:
		g.handleBind(intent, session, replyTo)

	case IntentApprove:
		g.handleApprovalResponse(intent, session, replyTo, msg)

	case IntentSendTask:
		g.handleSendTask(intent, session, replyTo)

	default:
		g.sendMessage(replyTo, &OutgoingMessage{
			Text: "I didn't understand that. Type `help` for available commands.",
		})
	}
}

// sendHelp sends help message.
func (g *Gateway) sendHelp(replyTo ReplyContext) {
	help := `**GoZen Bot Commands**

• ` + "`list`" + ` - List all connected processes
• ` + "`status [name]`" + ` - Show process status
• ` + "`bind <name>`" + ` - Bind to a process for subsequent commands
• ` + "`pause/resume/cancel [name]`" + ` - Control tasks
• ` + "`<name> <task>`" + ` - Send a task to a process

**Examples:**
• ` + "`status gozen`" + `
• ` + "`bind api`" + `
• ` + "`gozen run tests`" + `
• "帮我看看 gozen 的状态"`

	g.sendMessage(replyTo, &OutgoingMessage{
		Text:   help,
		Format: "markdown",
	})
}

// sendProcessList sends the list of connected processes.
func (g *Gateway) sendProcessList(replyTo ReplyContext) {
	processes := g.registry.List()

	if len(processes) == 0 {
		g.sendMessage(replyTo, &OutgoingMessage{
			Text: "No processes connected.",
		})
		return
	}

	var sb strings.Builder
	sb.WriteString("**Connected Processes**\n\n")

	for _, p := range processes {
		status := "🟢"
		if p.Status == "busy" {
			status = "🟡"
		} else if p.Status == "error" {
			status = "🔴"
		}

		name := p.Name
		if p.Alias != "" {
			name = fmt.Sprintf("%s (%s)", p.Alias, p.Name)
		}

		sb.WriteString(fmt.Sprintf("%s **%s**\n", status, name))
		sb.WriteString(fmt.Sprintf("   Path: `%s`\n", p.Path))
		if p.CurrentTask != "" {
			sb.WriteString(fmt.Sprintf("   Task: %s\n", p.CurrentTask))
		}
		sb.WriteString("\n")
	}

	g.sendMessage(replyTo, &OutgoingMessage{
		Text:   sb.String(),
		Format: "markdown",
	})
}

// handleStatusQuery handles status query intent.
func (g *Gateway) handleStatusQuery(intent *ParsedIntent, session *Session, replyTo ReplyContext) {
	target := intent.Target
	if target == "" {
		target = session.BoundProcess
	}

	// If still no target and only one process, use it
	if target == "" {
		processes := g.registry.List()
		if len(processes) == 1 {
			target = processes[0].Name
		} else if len(processes) == 0 {
			g.sendMessage(replyTo, &OutgoingMessage{Text: "No processes connected."})
			return
		} else {
			g.sendMessage(replyTo, &OutgoingMessage{
				Text: "Multiple processes available. Please specify which one or use `bind <name>` first.",
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

	// Build status message
	status := "🟢 Idle"
	if process.Status == "busy" {
		status = "🟡 Busy"
	} else if process.Status == "error" {
		status = "🔴 Error"
	}

	uptime := time.Since(process.StartTime).Round(time.Second)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("**%s** Status\n\n", process.Name))
	sb.WriteString(fmt.Sprintf("• Status: %s\n", status))
	sb.WriteString(fmt.Sprintf("• Path: `%s`\n", process.Path))
	sb.WriteString(fmt.Sprintf("• Uptime: %s\n", uptime))
	if process.CurrentTask != "" {
		sb.WriteString(fmt.Sprintf("• Task: %s\n", process.CurrentTask))
	}

	g.sendMessage(replyTo, &OutgoingMessage{
		Text:   sb.String(),
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
			g.sendMessage(replyTo, &OutgoingMessage{Text: "No processes connected."})
			return
		} else {
			g.sendMessage(replyTo, &OutgoingMessage{
				Text: "Multiple processes available. Please specify which one (e.g., `gozen run tests`).",
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
