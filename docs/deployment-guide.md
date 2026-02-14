# Deployment Guide: LiteLLM on HF Spaces + Model on HF Inference API

A step-by-step practical guide to deploy the full aiterm backend stack and test the app end-to-end.

**What you'll set up:**

```
aiterm CLI  ──►  LiteLLM Proxy (HF Spaces)  ──►  Model (HF Inference API)
```

**Time required:** ~15–20 minutes

---

## Prerequisites

- A [HuggingFace account](https://huggingface.co/join) (free)
- A HuggingFace API token with **write** access
- A free [Supabase](https://supabase.com) account (for the LiteLLM database)
- `aiterm` binary built and available (see main README)

---

## Step 1: Get Your HuggingFace Token

1. Go to [huggingface.co/settings/tokens](https://huggingface.co/settings/tokens)
2. Click **New token**
3. Name: `aiterm-deploy`
4. Type: **Write** (needed to create Spaces)
5. Click **Generate**
6. Copy the token — you'll need it in Steps 2 and 3

```
hf_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

---

## Step 2: Deploy the Model on HF Inference API

The HuggingFace Inference API lets you use models for free (rate-limited) or via paid endpoints.

### Option A: Use a Public Model (Free, No Setup)

HuggingFace hosts many models on their free Inference API. No deployment needed — just pick a model:

| Model | Good For | Model ID |
|-------|----------|----------|
| Qwen2.5-Coder-0.5B-Instruct | Fast, lightweight code/command generation | `Qwen/Qwen2.5-Coder-0.5B-Instruct` |
| Qwen2.5-Coder-1.5B-Instruct | Better quality, still fast | `Qwen/Qwen2.5-Coder-1.5B-Instruct` |
| CodeLlama-7b-Instruct-hf | Strong code generation | `codellama/CodeLlama-7b-Instruct-hf` |

**Test the model directly:**

```bash
curl https://api-inference.huggingface.co/models/Qwen/Qwen2.5-Coder-0.5B-Instruct/v1/chat/completions \
  -H "Authorization: Bearer hf_YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "Qwen/Qwen2.5-Coder-0.5B-Instruct",
    "messages": [
      {"role": "system", "content": "You are a shell command generator. Return ONLY the command, no explanation."},
      {"role": "user", "content": "Generate a single shell command for: list all files larger than 100MB"}
    ],
    "max_tokens": 200
  }'
```

If you get a valid response with a command, the model is ready. Move to Step 3.

### Option B: Create a Dedicated Inference Endpoint (Paid)

For production use with guaranteed availability and no rate limits:

1. Go to [ui.endpoints.huggingface.co](https://ui.endpoints.huggingface.co/)
2. Click **New Endpoint**
3. **Model**: `Qwen/Qwen2.5-Coder-0.5B-Instruct`
4. **Cloud**: AWS or Azure
5. **Instance**: CPU (cheapest) or GPU (faster)
6. **Scaling**: Enable **Scale to Zero** to save costs
7. Click **Create Endpoint**
8. Wait for status to show **Running**
9. Copy the endpoint URL:
   ```
   https://xxxx.aws.endpoints.huggingface.cloud
   ```

---

## Step 3: Deploy LiteLLM on HuggingFace Spaces

### 3.1 Create the Space

1. Go to [huggingface.co/new-space](https://huggingface.co/new-space)
2. Fill in:
   - **Space name**: `aiterm-proxy`
   - **SDK**: **Docker**
   - **Visibility**: **Private** (recommended)
3. Click **Create Space**

### 3.2 Set Secrets

Go to your Space → **Settings** → **Repository secrets** and add:

| Secret Name | Value | Required? |
|-------------|-------|-----------|
| `HF_TOKEN` | `hf_xxxxxxxx` (from Step 1) | **Yes** |
| `LITELLM_MASTER_KEY` | A strong random key (e.g. `openssl rand -hex 32`) | **Yes** |

> **Note:** A database is **not required** to get started. LiteLLM works without one — you just won't have virtual keys or spend tracking. See [Optional: Add a Database](#optional-add-a-database) later.

### 3.3 Upload Deploy Files

You need two files in the Space repo. Clone it and add them:

```bash
# Clone your Space
git clone https://huggingface.co/spaces/YOUR_USERNAME/aiterm-proxy
cd aiterm-proxy
```

**Create `config.yaml`:**

```yaml
model_list:
  # Option A: Free Inference API
  - model_name: "default"
    litellm_params:
      model: "huggingface/Qwen/Qwen2.5-Coder-0.5B-Instruct"
      api_key: "os.environ/HF_TOKEN"

  # Option B: Dedicated Endpoint (uncomment if using)
  # - model_name: "default"
  #   litellm_params:
  #     model: "huggingface/Qwen/Qwen2.5-Coder-0.5B-Instruct"
  #     api_key: "os.environ/HF_TOKEN"
  #     api_base: "https://your-endpoint.aws.endpoints.huggingface.cloud"

general_settings:
  master_key: "os.environ/LITELLM_MASTER_KEY"
  max_budget: 10.0
  budget_duration: "30d"

  # Uncomment these if you add a database later:
  # database_url: "os.environ/DATABASE_URL"
  # enable_user_auth: true
  # ui_access_mode: "admin_only"

rate_limits:
  - model: "default"
    tpm: 10000
    rpm: 60

litellm_settings:
  telemetry: false
  drop_params: true
  cache: true
  cache_params:
    type: "local"
    ttl: 600
```

**Create `Dockerfile`:**

```dockerfile
FROM docker.litellm.ai/berriai/litellm:main-stable

WORKDIR /app

COPY config.yaml .

EXPOSE 7860

# UI is available at /ui by default — no flag needed
CMD ["--port", "7860", "--config", "config.yaml"]
```

**Push to the Space:**

```bash
git add Dockerfile config.yaml
git commit -m "Deploy LiteLLM proxy"
git push
```

### 3.4 Wait for Build

1. Go to your Space page on HuggingFace
2. Watch the **Logs** tab — the build takes 2–5 minutes
3. When you see `LiteLLM Proxy started`, the proxy is live

Your proxy URL is:
```
https://YOUR_USERNAME-aiterm-proxy.hf.space
```

### 3.5 Verify the Proxy

```bash
# Health check
curl https://YOUR_USERNAME-aiterm-proxy.hf.space/health

# Expected: {"status":"healthy"}
```

Open the dashboard at:
```
https://YOUR_USERNAME-aiterm-proxy.hf.space/ui
```

Log in with your `LITELLM_MASTER_KEY`.

---

## Step 4: Get Your API Token

### Without a database (simplest)

Use your **master key** directly as the API token for aiterm. This works fine for personal use:

```bash
# Your LITELLM_MASTER_KEY is your API token
```

### With a database (virtual keys)

If you added a database (see [Optional: Add a Database](#optional-add-a-database)), you can create scoped virtual keys:

```bash
curl -X POST https://YOUR_USERNAME-aiterm-proxy.hf.space/key/generate \
  -H "Authorization: Bearer YOUR_MASTER_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "models": ["default"],
    "max_budget": 5.0,
    "budget_duration": "30d"
  }'
```

The response `key` field is your virtual key.

---

## Step 5: Configure aiterm

Now connect your aiterm CLI to the deployed proxy:

```bash
# Set the API endpoint to your LiteLLM proxy
aiterm config set api_endpoint https://YOUR_USERNAME-aiterm-proxy.hf.space/v1/chat/completions

# Set your virtual key
aiterm config set api_token sk-your-virtual-key-here

# Set the model name (must match model_name in config.yaml)
aiterm config set model default
```

Or run the setup wizard:

```bash
aiterm setup
# API Endpoint: https://YOUR_USERNAME-aiterm-proxy.hf.space/v1/chat/completions
# API Token: sk-your-virtual-key-here
# Model: default
```

Verify the config:

```bash
aiterm config
```

Expected output:

```
api_endpoint: https://YOUR_USERNAME-aiterm-proxy.hf.space/v1/chat/completions
api_token:    ***********xxxx
model:        default
shell:        auto
```

---

## Step 6: Test the App

### Basic Test

```bash
aiterm "list all files in the current directory"
```

Expected: prints a command like `ls -la` (Linux/macOS) or `Get-ChildItem` (Windows).

### Cross-Platform Test

```bash
aiterm "find all log files larger than 10MB" -t linux
# Expected: find / -name "*.log" -size +10M

aiterm "find all log files larger than 10MB" -t win
# Expected: Get-ChildItem -Path C:\ -Recurse -Filter *.log | Where-Object {$_.Length -gt 10MB}

aiterm "find all log files larger than 10MB" -t mac
# Expected: find / -name "*.log" -size +10M
```

### Headless / Pipe Test

```bash
aiterm generate "show current date and time"
# Prints just the command to stdout, e.g.: date
```

### Error Handling Tests

```bash
# Test with invalid token
aiterm config set api_token sk-invalid-key
aiterm "test"
# Expected: error about authentication

# Restore your key
aiterm config set api_token sk-your-virtual-key-here
```

---

## Troubleshooting

### "Connection refused" or timeout

- **Check the Space is running**: visit `https://YOUR_USERNAME-aiterm-proxy.hf.space`
- **HF Spaces may sleep** after inactivity — the first request wakes it up (~30s)
- **Check health**: `curl https://YOUR_USERNAME-aiterm-proxy.hf.space/health`

### "Authentication failed"

- If running **without a database**: use your `LITELLM_MASTER_KEY` as the `api_token`
- If running **with a database**: use a virtual key generated via the dashboard or API
- Verify the key matches exactly (no extra spaces)

### "Model not found" or empty response

- Check `config.yaml` — the `model_name` must match what you set in aiterm config
- Verify the HF model is accessible: test it with curl (Step 2)
- Check LiteLLM logs in the Space **Logs** tab

### "Database connection failed" (P1001 error)

This happens when `database_url` is set in `config.yaml` but the database is unreachable. Two fixes:

1. **Remove the database** (recommended for getting started): make sure `database_url` is commented out in `config.yaml`
2. **Fix the connection string**: ensure `DATABASE_URL` secret is a valid PostgreSQL URI:
   ```
   postgresql://postgres:PASSWORD@db.PROJECT.supabase.co:5432/postgres
   ```

### Slow responses

- First request may be slow (~10–30s) if the HF Space or model endpoint was sleeping
- Subsequent requests should be faster (1–5s)
- Consider enabling LiteLLM caching (already enabled in config)

---

## Optional: Add a Database

Adding PostgreSQL enables **virtual keys**, **spend tracking**, and **per-key budgets**. This is optional — LiteLLM works without it.

### Set Up Supabase (Free)

1. Go to [supabase.com](https://supabase.com) and sign in
2. Click **New Project**
   - Name: `litellm-aiterm`
   - Database password: choose a strong password (save it!)
   - Region: closest to your HF Space
3. Wait ~1 minute for creation
4. Go to **Project Settings** → **Database** → **Connection string** → **URI**
5. Copy the connection string:
   ```
   postgresql://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres
   ```

### Add to HF Space

1. Add a new secret in Space Settings:
   - `DATABASE_URL` = your Supabase connection string
2. Update `config.yaml` — uncomment the database lines:
   ```yaml
   general_settings:
     master_key: "os.environ/LITELLM_MASTER_KEY"
     database_url: "os.environ/DATABASE_URL"
     enable_user_auth: true
     ui_access_mode: "admin_only"
     max_budget: 10.0
     budget_duration: "30d"
   ```
3. Push and rebuild the Space

### Other Free Database Providers

| Provider | Free Tier |
|----------|-----------|
| [Supabase](https://supabase.com) | 500 MB |
| [Neon](https://neon.tech) | 512 MB |
| [Railway](https://railway.app) | Trial credits |

---

## Architecture Summary

```
┌──────────────────┐     HTTPS      ┌─────────────────────┐     HTTPS      ┌──────────────────┐
│                  │  ──────────►   │                     │  ──────────►   │                  │
│   aiterm CLI     │                │  LiteLLM Proxy      │                │  HF Inference    │
│   (your machine) │  ◄──────────   │  (HF Spaces)        │  ◄──────────   │  API / Endpoint  │
│                  │                │                     │                │                  │
└──────────────────┘                │  • Virtual keys     │                └──────────────────┘
                                    │  • Rate limiting    │
                                    │  • Spend tracking   │                ┌──────────────────┐
                                    │  • Response caching │                │  PostgreSQL      │
                                    │  • Admin dashboard  │  ◄──────────   │  (Supabase)      │
                                    └─────────────────────┘                └──────────────────┘
```

---

## Cost Breakdown

| Component | Cost |
|-----------|------|
| HuggingFace Inference API (free tier) | Free (rate-limited) |
| HuggingFace Space (Docker, private) | Free |
| Supabase (free tier, 500 MB) | Free |
| **Total** | **$0/month** |

For higher throughput, consider a Dedicated Inference Endpoint (~$0.06/hr for CPU).
