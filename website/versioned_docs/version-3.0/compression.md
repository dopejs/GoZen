---
sidebar_position: 14
title: Context Compression
---

# Context Compression

Automatically compress large conversation contexts to save tokens and costs.

## Overview

When conversations grow large, GoZen can:

1. Detect when context exceeds a threshold
2. Summarize older messages using a cheap model
3. Replace old messages with the summary
4. Forward the compressed request upstream

This is completely transparent to your CLI tools.

## Configuration

```json
{
  "compression": {
    "enabled": true,
    "threshold_tokens": 50000,
    "target_tokens": 20000,
    "summary_model": "claude-3-haiku-20240307",
    "preserve_recent": 4
  }
}
```

## Options

| Option | Default | Description |
|--------|---------|-------------|
| `enabled` | `false` | Enable compression (BETA) |
| `threshold_tokens` | `50000` | Trigger compression above this |
| `target_tokens` | `20000` | Compress to approximately this size |
| `summary_model` | `claude-3-haiku-20240307` | Model for summarization |
| `preserve_recent` | `4` | Keep last N messages uncompressed |
| `summary_provider` | (first healthy) | Provider for summarization |

## How It Works

1. **Detection**: GoZen estimates token count for each request
2. **Extraction**: Messages older than `preserve_recent` are extracted
3. **Summarization**: A cheap model summarizes the conversation history
4. **Injection**: Summary replaces old messages as a system message
5. **Forwarding**: Compressed request is sent to the provider

## Benefits

- **Cost Savings**: Reduce token usage by 50-80%
- **Avoid Limits**: Stay within context window limits
- **Transparent**: CLI tools don't need any changes
- **Preserves Context**: Important information is retained in summaries

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/v1/compression` | GET | Get compression config |
| `/api/v1/compression` | PUT | Update compression config |
| `/api/v1/compression/stats` | GET | Get compression statistics |

## Best Practices

1. Start with default thresholds and adjust based on your usage
2. Use a fast, cheap model for summarization (Haiku recommended)
3. Keep `preserve_recent` at 4+ to maintain conversation flow
4. Monitor compression stats to ensure quality
