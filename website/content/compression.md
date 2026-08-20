---
sidebar_position: 14
title: Context Compression (BETA)
---

# Context Compression (BETA)

:::warning BETA Feature
Context compression is currently in beta. It is disabled by default and requires explicit configuration to enable.
:::

Automatically compress conversation context when token count exceeds threshold, reducing costs while preserving conversation quality.

## Features

- **Automatic compression** — Triggered when token count exceeds threshold
- **Smart summarization** — Uses cheap model (claude-3-haiku) to summarize older messages
- **Recent message preservation** — Keeps recent messages intact for context continuity
- **Token estimation** — Accurate token counting before API calls
- **Statistics tracking** — Monitor compression effectiveness
- **Transparent operation** — Works seamlessly with all AI clients

## How It Works

1. **Token estimation** — Count tokens in conversation history
2. **Threshold check** — Compare against configured threshold (default: 50,000)
3. **Message selection** — Identify older messages for compression
4. **Summarization** — Use cheap model to create concise summary
5. **Context replacement** — Replace old messages with summary
6. **Request forwarding** — Send compressed context to target model

## Configuration

### Enable Compression

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 50000,
    "target_tokens": 20000,
    "summarizer_model": "claude-3-haiku-20240307",
    "preserve_recent_messages": 5,
    "tokens_per_char": 0.25
  }
}
```

**Options:**

| Option | Default | Description |
|--------|---------|-------------|
| `enabled` | `false` | Enable context compression |
| `threshold_tokens` | `50000` | Trigger compression when context exceeds this |
| `target_tokens` | `20000` | Target token count after compression |
| `summarizer_model` | `claude-3-haiku-20240307` | Model used for summarization |
| `preserve_recent_messages` | `5` | Number of recent messages to keep intact |
| `tokens_per_char` | `0.25` | Estimation ratio for token counting |

### Per-Profile Configuration

Enable compression for specific profiles:

```json
{
  "profiles": {
    "long-context": {
      "providers": ["anthropic"],
      "compression": {
        "enabled": true,
        "threshold_tokens": 100000,
        "target_tokens": 40000
      }
    },
    "short-context": {
      "providers": ["openai"],
      "compression": {
        "enabled": false
      }
    }
  }
}
```

## Token Estimation

GoZen uses character-based estimation for fast token counting:

```
estimated_tokens = character_count * tokens_per_char
```

**Default ratio:** 0.25 tokens per character (1 token ≈ 4 characters)

**Accuracy:** ±10% for English text, may vary for other languages

For exact token counting, GoZen uses the `tiktoken-go` library when available.

## Compression Strategy

### Message Selection

1. **System messages** — Always preserved
2. **Recent messages** — Last N messages preserved (default: 5)
3. **Older messages** — Candidates for compression

### Summarization Prompt

```
Summarize the following conversation history concisely while preserving key information, decisions, and context:

[older messages]

Provide a brief summary that captures the essential points.
```

### Result

```
Original: 45,000 tokens (30 messages)
After compression: 22,000 tokens (summary + 5 recent messages)
Savings: 23,000 tokens (51%)
```

## Web UI

Access compression settings at `http://localhost:19840/settings`:

1. Navigate to "Compression" tab (marked with BETA badge)
2. Toggle "Enable Compression"
3. Adjust threshold and target tokens
4. Select summarizer model
5. Set number of recent messages to preserve
6. Click "Save"

### Statistics Dashboard

View compression statistics:

- **Total compressions** — Number of times compression was triggered
- **Tokens saved** — Total tokens saved across all compressions
- **Average savings** — Average token reduction per compression
- **Compression rate** — Percentage of requests that triggered compression

## API Endpoints

### Get Compression Stats

```bash
GET /api/v1/compression/stats
```

Response:
```json
{
  "enabled": true,
  "total_compressions": 42,
  "tokens_saved": 1250000,
  "average_savings": 29761,
  "compression_rate": 0.15,
  "last_compression": "2026-03-05T10:30:00Z"
}
```

### Update Compression Settings

```bash
PUT /api/v1/compression/settings
Content-Type: application/json

{
  "enabled": true,
  "threshold_tokens": 60000,
  "target_tokens": 25000
}
```

### Reset Statistics

```bash
POST /api/v1/compression/stats/reset
```

## Use Cases

### Long Coding Sessions

**Scenario:** Multi-hour coding session with Claude Code

**Configuration:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 80000,
    "target_tokens": 30000,
    "preserve_recent_messages": 10
  }
}
```

**Benefit:** Maintain conversation continuity without hitting context limits

### Batch Processing

**Scenario:** Processing multiple documents with AI

**Configuration:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 40000,
    "target_tokens": 15000,
    "preserve_recent_messages": 3
  }
}
```

**Benefit:** Reduce costs while processing large document sets

### Research & Analysis

**Scenario:** Long research sessions with multiple topics

**Configuration:**
```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 100000,
    "target_tokens": 40000,
    "preserve_recent_messages": 8
  }
}
```

**Benefit:** Keep conversation focused on recent topics while preserving earlier context

## Best Practices

1. **Start with defaults** — Default settings work well for most use cases
2. **Monitor statistics** — Check compression rate and savings regularly
3. **Adjust threshold** — Increase for long-context models (Claude Opus), decrease for short-context
4. **Preserve enough messages** — Keep 5-10 recent messages for context continuity
5. **Use cheap summarizer** — Haiku is fast and cost-effective for summarization
6. **Test before production** — Verify compression quality with your specific use case

## Limitations

1. **Quality loss** — Summarization may lose nuanced details
2. **Latency increase** — Adds summarization API call overhead
3. **Cost trade-off** — Summarization costs vs. token savings
4. **Language support** — Works best with English, may vary for other languages
5. **Context window** — Cannot exceed model's maximum context window

## Troubleshooting

### Compression not triggering

1. Verify `compression.enabled` is `true`
2. Check token count exceeds threshold
3. Ensure conversation has enough messages to compress
4. Review daemon logs for compression errors

### Poor summarization quality

1. Try different summarizer model (e.g., claude-3-sonnet)
2. Increase `preserve_recent_messages` to keep more context
3. Adjust `target_tokens` to allow longer summaries
4. Check if summarizer model is available and working

### Increased latency

1. Compression adds one extra API call (summarization)
2. Use faster summarizer model (haiku is fastest)
3. Increase threshold to compress less frequently
4. Consider disabling for latency-sensitive applications

### Unexpected costs

1. Monitor summarization costs in usage dashboard
2. Compare savings vs. summarization costs
3. Adjust threshold to compress less frequently
4. Use cheapest available model for summarization

## Performance Impact

- **Token estimation** — ~1ms per request (negligible)
- **Summarization** — 1-3 seconds (depends on model and message count)
- **Memory overhead** — Minimal (~1KB per compression)
- **Cost savings** — Typically 30-50% token reduction

## Advanced Configuration

### Custom Summarization Prompt

```json
{
  "compression": {
    "enabled": true,
    "custom_prompt": "Create a technical summary of the following conversation, focusing on code changes, decisions, and action items:\n\n{messages}\n\nSummary:"
  }
}
```

### Conditional Compression

Enable compression only for specific scenarios:

```json
{
  "profiles": {
    "default": {
      "scenarios": {
        "longContext": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": true,
            "threshold_tokens": 100000
          }
        },
        "default": {
          "providers": ["anthropic"],
          "compression": {
            "enabled": false
          }
        }
      }
    }
  }
}
```

### Multi-Stage Compression

Compress multiple times for very long conversations:

```json
{
  "compression": {
    "enabled": true,
    "stages": [
      {
        "threshold_tokens": 50000,
        "target_tokens": 30000
      },
      {
        "threshold_tokens": 80000,
        "target_tokens": 40000
      }
    ]
  }
}
```

## Future Enhancements

- Semantic similarity matching for intelligent message selection
- Multi-model summarization for quality comparison
- Compression quality metrics and feedback
- Custom compression strategies per use case
- Integration with RAG for external context storage
