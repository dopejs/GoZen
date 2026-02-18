package agent

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dopejs/gozen/internal/config"
)

// Runtime manages autonomous agent task execution.
type Runtime struct {
	config    *config.RuntimeConfig
	tasks     map[string]*RuntimeTask
	client    *http.Client
	proxyPort int
	mu        sync.RWMutex
}

// Global runtime instance
var (
	globalRuntime     *Runtime
	globalRuntimeOnce sync.Once
	globalRuntimeMu   sync.RWMutex
)

// InitGlobalRuntime initializes the global runtime.
func InitGlobalRuntime(proxyPort int) {
	globalRuntimeOnce.Do(func() {
		cfg := config.GetAgent()
		var rtCfg *config.RuntimeConfig
		if cfg != nil {
			rtCfg = cfg.Runtime
		}
		globalRuntimeMu.Lock()
		globalRuntime = NewRuntime(rtCfg, proxyPort)
		globalRuntimeMu.Unlock()
	})
}

// GetGlobalRuntime returns the global runtime.
func GetGlobalRuntime() *Runtime {
	globalRuntimeMu.RLock()
	defer globalRuntimeMu.RUnlock()
	return globalRuntime
}

// NewRuntime creates a new runtime.
func NewRuntime(cfg *config.RuntimeConfig, proxyPort int) *Runtime {
	if cfg == nil {
		cfg = &config.RuntimeConfig{
			Enabled:         false,
			PlanningModel:   "claude-sonnet-4-20250514",
			ExecutionModel:  "claude-sonnet-4-20250514",
			ValidationModel: "claude-haiku-3-5-20241022",
			MaxTurns:        50,
			MaxTokens:       500000,
		}
	}
	if cfg.PlanningModel == "" {
		cfg.PlanningModel = "claude-sonnet-4-20250514"
	}
	if cfg.ExecutionModel == "" {
		cfg.ExecutionModel = "claude-sonnet-4-20250514"
	}
	if cfg.ValidationModel == "" {
		cfg.ValidationModel = "claude-haiku-3-5-20241022"
	}
	if cfg.MaxTurns == 0 {
		cfg.MaxTurns = 50
	}
	if cfg.MaxTokens == 0 {
		cfg.MaxTokens = 500000
	}

	return &Runtime{
		config:    cfg,
		tasks:     make(map[string]*RuntimeTask),
		proxyPort: proxyPort,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// IsEnabled returns whether the runtime is enabled.
func (r *Runtime) IsEnabled() bool {
	return r.config != nil && r.config.Enabled
}

// UpdateConfig updates the runtime configuration.
func (r *Runtime) UpdateConfig(cfg *config.RuntimeConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.config = cfg
}

// StartTask starts a new autonomous task.
func (r *Runtime) StartTask(description string) (*RuntimeTask, error) {
	if !r.IsEnabled() {
		return nil, fmt.Errorf("runtime is not enabled")
	}

	task := &RuntimeTask{
		ID:          generateRuntimeTaskID(),
		Description: description,
		Status:      RuntimeStatusPlanning,
		CreatedAt:   time.Now(),
		StartedAt:   time.Now(),
		Turns:       make([]*AgentTurn, 0),
	}

	r.mu.Lock()
	r.tasks[task.ID] = task
	r.mu.Unlock()

	// Start execution in background
	go r.executeTask(task)

	return task, nil
}

// GetTask returns a task by ID.
func (r *Runtime) GetTask(id string) *RuntimeTask {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tasks[id]
}

// GetAllTasks returns all tasks.
func (r *Runtime) GetAllTasks() []*RuntimeTask {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*RuntimeTask, 0, len(r.tasks))
	for _, t := range r.tasks {
		tasks = append(tasks, t)
	}
	return tasks
}

// CancelTask cancels a running task.
func (r *Runtime) CancelTask(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	task, ok := r.tasks[id]
	if !ok {
		return false
	}

	if task.Status == RuntimeStatusPlanning || task.Status == RuntimeStatusExecuting || task.Status == RuntimeStatusValidating {
		task.Status = RuntimeStatusCancelled
		task.CompletedAt = time.Now()
		return true
	}
	return false
}

// executeTask runs the autonomous task execution loop.
func (r *Runtime) executeTask(task *RuntimeTask) {
	defer func() {
		if rec := recover(); rec != nil {
			r.mu.Lock()
			task.Status = RuntimeStatusFailed
			task.Result = &TaskResult{
				Success: false,
				Error:   fmt.Sprintf("panic: %v", rec),
			}
			task.CompletedAt = time.Now()
			r.mu.Unlock()
		}
	}()

	// Phase 1: Planning
	plan, err := r.planTask(task)
	if err != nil {
		r.failTask(task, err)
		return
	}
	task.Plan = plan

	// Check if cancelled
	if r.isTaskCancelled(task.ID) {
		return
	}

	// Phase 2: Execution
	r.mu.Lock()
	task.Status = RuntimeStatusExecuting
	r.mu.Unlock()

	var lastOutput string
	for i, step := range plan.Steps {
		if r.isTaskCancelled(task.ID) {
			return
		}

		// Check limits
		if len(task.Turns) >= r.config.MaxTurns || task.TotalTokens >= r.config.MaxTokens {
			r.failTask(task, fmt.Errorf("task limits exceeded"))
			return
		}

		r.mu.Lock()
		task.Plan.CurrentStep = i
		r.mu.Unlock()

		output, err := r.executeStep(task, step, lastOutput)
		if err != nil {
			r.failTask(task, err)
			return
		}
		lastOutput = output
	}

	// Check if cancelled
	if r.isTaskCancelled(task.ID) {
		return
	}

	// Phase 3: Validation
	r.mu.Lock()
	task.Status = RuntimeStatusValidating
	r.mu.Unlock()

	valid, err := r.validateResult(task, lastOutput)
	if err != nil {
		r.failTask(task, err)
		return
	}

	// Complete task
	r.mu.Lock()
	task.Status = RuntimeStatusCompleted
	task.CompletedAt = time.Now()
	task.Result = &TaskResult{
		Success: valid,
		Output:  lastOutput,
		Tokens:  task.TotalTokens,
		Cost:    task.TotalCost,
	}
	r.mu.Unlock()
}

// planTask creates an execution plan for the task.
func (r *Runtime) planTask(task *RuntimeTask) (*TaskPlan, error) {
	prompt := fmt.Sprintf(`You are a task planner. Break down the following task into clear, actionable steps.
Each step should be specific and achievable in a single action.
Return ONLY a JSON array of step descriptions, nothing else.

Task: %s

Example output format:
["Step 1 description", "Step 2 description", "Step 3 description"]`, task.Description)

	response, tokens, cost, err := r.sendRequest(task, r.config.PlanningModel, "planning", prompt)
	if err != nil {
		return nil, fmt.Errorf("planning failed: %w", err)
	}

	r.mu.Lock()
	task.TotalTokens += tokens
	task.TotalCost += cost
	r.mu.Unlock()

	// Parse steps from response
	var steps []string
	// Try to extract JSON array from response
	response = strings.TrimSpace(response)
	if err := json.Unmarshal([]byte(response), &steps); err != nil {
		// Try to find JSON in response
		start := strings.Index(response, "[")
		end := strings.LastIndex(response, "]")
		if start >= 0 && end > start {
			if err := json.Unmarshal([]byte(response[start:end+1]), &steps); err != nil {
				// Fallback: treat entire response as single step
				steps = []string{task.Description}
			}
		} else {
			steps = []string{task.Description}
		}
	}

	if len(steps) == 0 {
		steps = []string{task.Description}
	}

	return &TaskPlan{
		Steps:       steps,
		CurrentStep: 0,
	}, nil
}

// executeStep executes a single step of the plan.
func (r *Runtime) executeStep(task *RuntimeTask, step, previousOutput string) (string, error) {
	var prompt string
	if previousOutput != "" {
		prompt = fmt.Sprintf(`You are executing a task step by step.

Previous output:
%s

Current step to execute:
%s

Execute this step and provide the result.`, previousOutput, step)
	} else {
		prompt = fmt.Sprintf(`You are executing a task step by step.

Current step to execute:
%s

Execute this step and provide the result.`, step)
	}

	response, tokens, cost, err := r.sendRequest(task, r.config.ExecutionModel, "execution", prompt)
	if err != nil {
		return "", fmt.Errorf("execution failed: %w", err)
	}

	r.mu.Lock()
	task.TotalTokens += tokens
	task.TotalCost += cost
	r.mu.Unlock()

	return response, nil
}

// validateResult validates the task result.
func (r *Runtime) validateResult(task *RuntimeTask, output string) (bool, error) {
	prompt := fmt.Sprintf(`You are a task validator. Review the following task and its output.
Determine if the task was completed successfully.

Original task: %s

Output:
%s

Respond with ONLY "VALID" if the task was completed successfully, or "INVALID: <reason>" if not.`, task.Description, output)

	response, tokens, cost, err := r.sendRequest(task, r.config.ValidationModel, "validation", prompt)
	if err != nil {
		return false, fmt.Errorf("validation failed: %w", err)
	}

	r.mu.Lock()
	task.TotalTokens += tokens
	task.TotalCost += cost
	r.mu.Unlock()

	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(response)), "VALID"), nil
}

// sendRequest sends a request to the AI model.
func (r *Runtime) sendRequest(task *RuntimeTask, model, phase, prompt string) (string, int, float64, error) {
	reqBody := map[string]interface{}{
		"model":      model,
		"max_tokens": 4096,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	reqData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, 0, err
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/default/runtime-%s/v1/messages", r.proxyPort, task.ID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(reqData))
	if err != nil {
		return "", 0, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := r.client.Do(req)
	if err != nil {
		return "", 0, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", 0, 0, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var respData struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &respData); err != nil {
		return "", 0, 0, err
	}

	// Record turn
	turn := &AgentTurn{
		Model:     model,
		Phase:     phase,
		Request:   reqData,
		Response:  body,
		Tokens:    respData.Usage.InputTokens + respData.Usage.OutputTokens,
		Timestamp: time.Now(),
	}

	r.mu.Lock()
	task.Turns = append(task.Turns, turn)
	r.mu.Unlock()

	// Extract text
	if len(respData.Content) == 0 {
		return "", 0, 0, fmt.Errorf("empty response")
	}

	// Calculate cost (simplified)
	cost := float64(respData.Usage.InputTokens)*0.003/1000 + float64(respData.Usage.OutputTokens)*0.015/1000

	return respData.Content[0].Text, respData.Usage.InputTokens + respData.Usage.OutputTokens, cost, nil
}

// failTask marks a task as failed.
func (r *Runtime) failTask(task *RuntimeTask, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	task.Status = RuntimeStatusFailed
	task.CompletedAt = time.Now()
	task.Result = &TaskResult{
		Success: false,
		Error:   err.Error(),
		Tokens:  task.TotalTokens,
		Cost:    task.TotalCost,
	}
}

// isTaskCancelled checks if a task has been cancelled.
func (r *Runtime) isTaskCancelled(id string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return true
	}
	return task.Status == RuntimeStatusCancelled
}

// generateRuntimeTaskID generates a unique runtime task ID.
func generateRuntimeTaskID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "rt-" + hex.EncodeToString(b)
}
