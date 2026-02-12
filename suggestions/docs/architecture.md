# System Architecture

## Overview

AI Terminal Pro uses a layered architecture separating concerns across multiple services.

## Layers

### Layer 1: Terminal Application (Go + Wails)

**Responsibilities:**
- PTY management for real shell processes
- AI client for LiteLLM API
- Command validation and safety
- Configuration management

**Key Components:**
- `terminal/pty.go` - Cross-platform PTY management
- `ai/client.go` - LiteLLM API integration
- `security/validator.go` - Command risk analysis
- `config/settings.go` - User preferences

### Layer 2: LiteLLM Proxy (HuggingFace Spaces)

**Responsibilities:**
- Virtual key authentication
- Request routing to model endpoints
- Cost tracking and rate limiting
- Fallback to local models

**Technology:**
- LiteLLM proxy server
- Docker deployment on HF Spaces
- Supabase PostgreSQL for persistence

### Layer 3: Model Endpoints

**Primary:** HuggingFace Dedicated Endpoint
- Qwen3-0.6B fine-tuned model
- Scale-to-zero for cost savings
- ~15s cold start

**Fallback:** Local GGUF
- llama.cpp runtime
- Q4_K_M quantized model
- Zero latency, offline capable

## Data Flow

```
1. User opens terminal app
2. App sends warmup ping to LiteLLM (background)
3. LiteLLM wakes HF endpoint (if sleeping)
4. User presses Ctrl+K for AI mode
5. App sends context to LiteLLM
6. LiteLLM validates virtual key
7. LiteLLM routes to HF endpoint
8. Model generates command
9. App validates command safety
10. User reviews and executes
```

## Security Model

### Authentication Flow

1. User enters LiteLLM virtual key (NOT HF token)
2. Key validated against Supabase
3. Budget and rate limits checked
4. Request proxied with internal HF token
5. Response returned to user

### Command Safety

**Validation Pipeline:**
1. Parse command structure
2. Check against risk database
3. Normalize and validate
4. User confirmation for risky commands
5. Execute in isolated PTY

## Cost Optimization

**Strategies:**
- Scale-to-zero on HF endpoint
- Local GGUF fallback for common commands
- LiteLLM caching for identical prompts
- Virtual key budgets per user

## Monitoring

**Metrics:**
- Request latency (TTFT)
- Token generation speed
- Cost per request
- Error rates
- Active users
