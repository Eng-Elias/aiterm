# LiteLLM Proxy Deployment

Deploy a [LiteLLM](https://docs.litellm.ai/) proxy as the AI backend for **aiterm**. LiteLLM provides virtual key management, rate limiting, spend tracking, and unified routing to any LLM provider.

---

## Deployment Options

| Method | Best For | Port |
|--------|----------|------|
| [Docker Compose (local)](#docker-compose-local) | Development, self-hosted | 4000 |
| [HuggingFace Spaces](#huggingface-spaces) | Free cloud hosting | 7860 |

---

## Docker Compose (Local)

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- A HuggingFace API token ([create one](https://huggingface.co/settings/tokens))

### Quick Start

```bash
cd deploy/litellm

# 1. Create your environment file
cp .env.example .env

# 2. Edit .env — set HF_TOKEN and LITELLM_MASTER_KEY
#    Generate a master key:  openssl rand -hex 32

# 3. Start the stack
docker compose up -d

# 4. Verify
curl http://localhost:4000/health
```

### What Gets Deployed

| Container | Purpose |
|-----------|---------|
| `litellm-proxy` | LiteLLM proxy server (API + dashboard) |
| `litellm-db` | PostgreSQL 16 for virtual keys & spend tracking |

### Dashboard

Open **http://localhost:4000/ui** and log in with your `LITELLM_MASTER_KEY`.

From the dashboard you can:
- Create and manage virtual keys
- Monitor spend and usage
- Set per-key budgets and rate limits

### Create a Virtual Key

```bash
curl -X POST http://localhost:4000/key/generate \
  -H "Authorization: Bearer $LITELLM_MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "models": ["default"],
    "max_budget": 5.0,
    "budget_duration": "30d"
  }'
```

The response contains a `key` field — use this as your `api_token` in aiterm:

```bash
aiterm config set api_token sk-returned-virtual-key
aiterm config set api_endpoint http://localhost:4000/v1/chat/completions
```

### Stop / Restart

```bash
docker compose down          # stop
docker compose up -d         # start
docker compose logs -f       # view logs
docker compose down -v       # stop and delete database volume
```

---

## HuggingFace Spaces

Deploy LiteLLM for free on HuggingFace Spaces (Docker SDK).

### Steps

1. **Create a new Space** at [huggingface.co/new-space](https://huggingface.co/new-space)
   - SDK: **Docker**
   - Visibility: **Private** (recommended)

2. **Upload files** to the Space repo:
   - `Dockerfile.hf-spaces` → rename to `Dockerfile`
   - `config.yaml`

3. **Set Secrets** in Space Settings → Repository secrets:

   | Secret | Value |
   |--------|-------|
   | `HF_TOKEN` | Your HuggingFace token |
   | `LITELLM_MASTER_KEY` | Strong random key |
   | `DATABASE_URL` | PostgreSQL connection string (e.g. Supabase) |
   | `UI_PASSWORD` | Dashboard admin password |

4. **Get your endpoint URL:**
   ```
   API:       https://<user>-<space>.hf.space/v1/chat/completions
   Dashboard: https://<user>-<space>.hf.space/ui
   ```

5. **Configure aiterm:**
   ```bash
   aiterm setup
   # API Endpoint: https://<user>-<space>.hf.space/v1/chat/completions
   # API Token: your virtual key from the dashboard
   ```

### Database for HF Spaces

HF Spaces don't include a database. Use one of:

| Provider | Free Tier | Setup |
|----------|-----------|-------|
| [Supabase](https://supabase.com) | 500 MB | Create project → Settings → Database → Connection string |
| [Neon](https://neon.tech) | 512 MB | Create project → Connection Details |
| [Railway](https://railway.app) | Trial | New project → Add PostgreSQL |

---

## Configure aiterm to Use LiteLLM

After deploying, point aiterm at your proxy:

```bash
# Local Docker
aiterm config set api_endpoint http://localhost:4000/v1/chat/completions
aiterm config set api_token sk-your-virtual-key

# HuggingFace Spaces
aiterm config set api_endpoint https://<user>-<space>.hf.space/v1/chat/completions
aiterm config set api_token sk-your-virtual-key

# Set the model name (must match a model_name in config.yaml)
aiterm config set model default
```

---

## Customizing Models

Edit `config.yaml` to add or change models:

```yaml
model_list:
  # HuggingFace Inference API
  - model_name: "hf-model"
    litellm_params:
      model: "huggingface/Qwen/Qwen2.5-Coder-0.5B-Instruct"
      api_key: "os.environ/HF_TOKEN"

  # OpenAI
  - model_name: "openai"
    litellm_params:
      model: "openai/gpt-4o-mini"
      api_key: "os.environ/OPENAI_API_KEY"

  # Ollama (local)
  - model_name: "local"
    litellm_params:
      model: "ollama/llama3"
      api_base: "http://host.docker.internal:11434"
```

After editing, restart the proxy:

```bash
docker compose restart litellm
```

---

## Troubleshooting

### "Connection refused" from aiterm

- Verify the proxy is running: `docker compose ps`
- Check health: `curl http://localhost:4000/health`
- Check logs: `docker compose logs litellm`

### "Authentication failed"

- Ensure you're using a **virtual key** (not the master key) for aiterm
- Verify the key is active in the dashboard

### Database connection errors

- Check `DATABASE_URL` format in `.env`
- For local Docker: use `postgresql://litellm:litellm@db:5432/litellm`
- For external DB: ensure the host is reachable from the container

### HF Spaces shows "Building"

- Check Space logs for build errors
- Ensure all secrets are set correctly
- Verify `Dockerfile` (not `Dockerfile.hf-spaces`) is in the Space repo root
