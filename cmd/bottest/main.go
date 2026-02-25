package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dopejs/gozen/internal/bot"
	"github.com/dopejs/gozen/internal/config"
)

func main() {
	fmt.Println("Bot Chat Test Harness")
	fmt.Println("Commands: quit, clear")
	fmt.Println("---")

	store := config.DefaultStore()
	botCfg := store.GetBot()
	if botCfg == nil || botCfg.Profile == "" {
		fmt.Println("Error: Bot not configured. Set a profile in the Web UI first.")
		os.Exit(1)
	}

	proxyPort := store.GetProxyPort()
	if proxyPort == 0 {
		proxyPort = 19841
	}

	llm := bot.NewLLMClient(proxyPort, botCfg.Profile, botCfg.Model)

	memory, _ := bot.LoadMemory(bot.MemoryDir())
	systemPrompt := bot.BuildSystemPrompt(nil, botCfg.Profile, memory)

	var history []bot.ChatMessage
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nYou: ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch input {
		case "quit", "exit":
			fmt.Println("Goodbye!")
			return
		case "clear":
			history = nil
			fmt.Println("History cleared.")
			continue
		}

		history = append(history, bot.ChatMessage{Role: "user", Content: input})

		fmt.Print("\nBot: ")
		var response strings.Builder

		err := llm.ChatStream(context.Background(), systemPrompt, history, func(delta string) {
			fmt.Print(delta)
			response.WriteString(delta)
		})

		if err != nil {
			fmt.Printf("\nError: %v\n", err)
			// Remove failed user message
			history = history[:len(history)-1]
			continue
		}

		fmt.Println()
		history = append(history, bot.ChatMessage{Role: "assistant", Content: response.String()})
	}
}

// httpChat uses the HTTP API instead of direct LLM calls (alternative implementation)
func httpChat(webPort int, sessionID, message string) (string, string, error) {
	reqBody, _ := json.Marshal(map[string]string{
		"message":    message,
		"session_id": sessionID,
	})

	resp, err := http.Post(
		fmt.Sprintf("http://127.0.0.1:%d/api/v1/bot/chat", webPort),
		"application/json",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		return "", sessionID, err
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	var content string
	newSessionID := sessionID

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			event := strings.TrimPrefix(line, "event: ")
			if scanner.Scan() {
				dataLine := scanner.Text()
				if strings.HasPrefix(dataLine, "data: ") {
					data := strings.TrimPrefix(dataLine, "data: ")
					var payload map[string]string
					json.Unmarshal([]byte(data), &payload)

					switch event {
					case "session":
						newSessionID = payload["session_id"]
					case "delta":
						fmt.Print(payload["content"])
						content += payload["content"]
					case "done":
						content = payload["content"]
					case "error":
						return "", newSessionID, fmt.Errorf("%s", payload["error"])
					}
				}
			}
		}
	}

	return content, newSessionID, nil
}
