---
title: AI Terminal Pro
emoji: 🦅
colorFrom: blue
colorTo: green
sdk: docker
app_port: 7860
---

# AI Terminal - LiteLLM Proxy

This Space hosts the LiteLLM proxy for the AI Terminal application.

## Dashboard Access

**URL:** https://huggingface.co/spaces/{username}/{space_name}/ui

## API Endpoint

**Base URL:** https://huggingface.co/spaces/{username}/{space_name}

## Environment Variables

Set these in Space Settings > Secrets:
- `HF_TOKEN`: Your HuggingFace token
- `LITELLM_MASTER_KEY`: Admin key for dashboard
- `DATABASE_URL`: Supabase PostgreSQL connection
- `UI_PASSWORD`: Dashboard admin password
