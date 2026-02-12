# Security Model

## Overview

Security is implemented through layered controls protecting both the system and end users.

## Virtual Key System

### Why Not Raw HF Tokens?

**Raw Token Risks:**
- Irrevocable (can't disable without regenerating)
- Full access (no scoping)
- No usage tracking
- Exposed in settings/logs

**Virtual Key Benefits:**
- ✅ Revocable (enable/disable instantly)
- ✅ Scoped (specific models only)
- ✅ Tracked (cost per key)
- ✅ Budget-limited (prevent overage)
- ✅ Rate-limited (prevent abuse)

### Key Structure

```
sk-litellm-{tier}-{random}

Examples:
- sk-litellm-free-a1b2c3d4
- sk-litellm-pro-e5f6g7h8
- sk-litellm-ent-i9j0k1l2
```

### Storage

**OS-Specific Secure Storage:**
- Windows: Windows Credential Manager
- macOS: Keychain
- Linux: libsecret / Secret Service

**Never:**
- Store in plain text files
- Log to console
- Include in error messages
- Expose in UI

## Command Validation

### Risk Levels

| Level | Examples | Action |
|-------|----------|--------|
| None | ls, cd, pwd | Execute |
| Low | sudo apt-get | Warning |
| Medium | chmod 777 | Confirm |
| High | curl \| bash | Block + Explain |
| Critical | rm -rf / | Hard block |

### Blocked Patterns

**Always Rejected:**
- `rm -rf /` or variations
- `:(){ :|:& };:` (fork bomb)
- `curl ... | bash` (remote execution)
- `wget ... | sh`
- `eval(...)` with user input
- Format commands on system disks

**Require Confirmation:**
- Recursive chmod/chown
- Writing to /etc, /bin, /usr
- Network downloads
- Package installations
- Service restarts

## Data Privacy

### What We Log

✅ **Logged:**
- Command generated (after validation)
- Success/failure status
- Latency metrics
- Cost per request
- Virtual key used (hashed)

❌ **Never Logged:**
- User's natural language prompts
- Command output (may contain secrets)
- Environment variables
- File contents
- Raw API keys

**Retention:** 30 days, then auto-delete

### PII Protection

**Personal Identifiable Information:**
- Prompts may contain usernames, paths, project names
- These are processed but not stored
- Only command and metadata retained

## Rate Limiting

### Per-Key Limits

```yaml
Free Tier:
  rpm: 30
  tpm: 5000
  
Pro Tier:
  rpm: 120
  tpm: 20000
  
Enterprise:
  rpm: 600
  tpm: 100000
```

### Purpose

- Prevent abuse and DDoS
- Control costs per user
- Ensure fair resource sharing

## Access Control

### Roles

| Role | Endpoint Access | Dashboard | Admin |
|------|-----------------|-----------|-------|
| End User | Virtual Key only | No | No |
| Pro User | Virtual Key | Read-only | No |
| Admin | Master Key | Full | Yes |

### Admin Capabilities

- Enable/disable virtual keys
- View all spend logs
- Set global rate limits
- Manage model access
- Configure budgets

## Incident Response

### Key Compromise

1. Immediately disable virtual key
2. Notify affected user
3. Generate new key
4. Review logs for abuse
5. Document incident

### API Abuse

1. Rate limit triggered
2. Automatic temporary block
3. Alert admin if repeated
4. Review pattern for attack

## Compliance

### GDPR / Data Protection

- User can request data deletion
- 30-day retention limit
- No PII in logs
- Opt-out telemetry available

### Audit Trail

All administrative actions logged:
- Key creation/deletion
- Budget changes
- Rate limit updates
- Access denials
