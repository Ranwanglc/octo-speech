# OctoSpeech

Intelligent speech-to-text microservice for [OCTO](https://github.com/Mininglamp-OSS) — multi-engine ASR with context-aware correction, vocabulary management, and a built-in admin console.

## Features

- **Multi-Engine ASR** — Gemini, GPT, Qwen with automatic failover
- **Context-Aware Correction** — vocabulary profiles scoped per user/space/org
- **Local ASR Fallback** — optional on-device transcription for low-latency scenarios
- **Admin Console** — standalone management service with web UI for app/key lifecycle
- **SHA-256 Key Storage** — API keys stored as hashes, shown only once on creation
- **Audit Logging** — all admin operations tracked for compliance

## Architecture

```
┌─────────────────┐     ┌─────────────────┐
│  octo-speech    │     │  octo-speech    │
│  (Speech API)   │     │  (Admin)        │
│  :8780          │     │  :8781          │
└────────┬────────┘     └────────┬────────┘
         │                       │
         └───────────┬───────────┘
                     │
              ┌──────┴──────┐
              │    MySQL    │
              └─────────────┘
```

Two independent services sharing one database:

| Service | Port | Purpose |
|---------|------|---------|
| **speech** | 8780 | Transcription API, vocabulary management, config |
| **admin** | 8781 | App CRUD, API key management, web UI |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- MySQL 8.0+

### 1. Clone

```bash
git clone https://github.com/Mininglamp-OSS/octo-speech.git
cd octo-speech
```

### 2. Configure

Create a `.env` file:

```env
# MySQL
MYSQL_ROOT_PASSWORD=your_password

# Speech Service
SPEECH_DB_DSN=root:your_password@tcp(mysql:3306)/octo_speech?parseTime=true&loc=Asia%2FShanghai
VOICE_LITELLM_URL=https://your-llm-gateway/v1
VOICE_LITELLM_KEY=sk-xxx
VOICE_ENGINE=gemini

# Admin Service
ADMIN_USERNAME=admin
ADMIN_PASSWORD=your_admin_password
ADMIN_JWT_SECRET=your_secret_here
ADMIN_SECURE_COOKIE=false   # set true behind HTTPS
```

### 3. Deploy

```bash
# Build
docker build -t octo-speech:latest .
docker build -f Dockerfile.admin -t octo-speech-admin:latest .

# Run
docker compose up -d
```

### 4. Create Your First App

Open `http://localhost:8781` → Login → Create App → Copy the API key (shown only once).

## Speech API

All endpoints require `Authorization: Bearer <api_key>`.

### POST /v1/speech/transcribe

Transcribe audio with context-aware correction.

```bash
curl -X POST http://localhost:8780/v1/speech/transcribe \
  -H "Authorization: Bearer sk-xxx" \
  -F "audio=@recording.wav" \
  -F "context_text=previous message text" \
  -F "engine=gemini"
```

**Parameters:**

| Field | Type | Description |
|-------|------|-------------|
| `audio` | file | Audio file (max 5MB) |
| `context_text` | string | Preceding text for edit-mode |
| `chat_context` | string | Recent chat messages for context |
| `personal_context` | string | User's vocabulary profile |
| `member_context` | string | Group member names for recognition |
| `engine` | string | `gemini` / `gpt` / `qwen` |
| `model` | string | Specific model override |
| `mode` | string | `smart` / `append_only` / `edit_only` |
| `channel_type` | string | `dm` / `group` |

### GET /v1/speech/config

Returns service configuration (engine, limits, local ASR settings).

### PUT /v1/speech/vocabularies

Create or update a vocabulary correction profile.

### GET /v1/speech/vocabularies

Retrieve vocabulary profile with scope priority resolution.

### DELETE /v1/speech/vocabularies

Remove a vocabulary profile.

## Admin API

Admin service runs on port 8781 with JWT cookie authentication.

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/login` | Login (sets httpOnly cookie) |
| POST | `/api/logout` | Logout (clears cookies) |
| GET | `/api/apps` | List all apps |
| POST | `/api/apps` | Create app (returns key once) |
| PUT | `/api/apps/:id/status` | Enable/disable app |
| DELETE | `/api/apps/:id` | Delete app + related data |
| POST | `/api/apps/:id/reset-key` | Reset API key |
| GET | `/healthz` | Health check |

## Configuration

### Speech Service Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SPEECH_DB_DSN` | — | MySQL connection string (required) |
| `SPEECH_SERVICE_PORT` | `8780` | Listen port |
| `SPEECH_APP_CACHE_TTL` | `60` | Auth cache TTL in seconds |
| `VOICE_LITELLM_URL` | — | LLM gateway URL (required) |
| `VOICE_LITELLM_KEY` | — | LLM gateway API key (required) |
| `VOICE_ENGINE` | `gemini` | Default engine: gemini/gpt/qwen |
| `VOICE_MODELS` | gemini-3.1-pro-preview,... | Gemini model list |
| `VOICE_GPT_MODELS` | gpt-4o-mini-transcribe | GPT model list |
| `VOICE_QWEN_MODELS` | qwen3.5-omni-plus | Qwen model list |
| `VOICE_MAX_DURATION` | `60` | Max audio duration (seconds) |
| `VOICE_MAX_FILE_SIZE` | `3145728` | Max upload size (bytes) |
| `SPEECH_READ_TIMEOUT` | `30` | HTTP read timeout (seconds) |
| `SPEECH_WRITE_TIMEOUT` | `60` | HTTP write timeout (seconds) |
| `SPEECH_IDLE_TIMEOUT` | `120` | HTTP idle timeout (seconds) |

### Admin Service Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SPEECH_DB_DSN` | — | MySQL connection string (required) |
| `ADMIN_PORT` | `8781` | Listen port |
| `ADMIN_USERNAME` | — | Login username (required) |
| `ADMIN_PASSWORD` | — | Login password (required) |
| `ADMIN_JWT_SECRET` | random | JWT signing key (set for production) |
| `ADMIN_TOKEN_EXPIRE` | `24` | JWT lifetime in hours |
| `ADMIN_SECURE_COOKIE` | `true` | Require HTTPS for cookies |
| `ADMIN_TRUSTED_PROXIES` | — | Trusted proxy IPs for rate limiting |

## Security

- API keys stored as SHA-256 hashes — database breach doesn't expose keys
- Admin passwords bcrypt-hashed at startup — plaintext never held in memory
- JWT in httpOnly + SameSite=Strict cookies — XSS-resistant
- CSRF double-submit cookie pattern on all mutating endpoints
- Login rate limiting (5/min/IP) with trusted proxy support
- Non-root container runtime (UID 10001)
- Configurable HTTP server timeouts against slowloris
- Request body size enforcement (5MB) before multipart parsing
- Model parameter whitelist validation

## Development

```bash
# Run tests
go test ./...

# Build locally
go build -o speech ./cmd/speech
go build -o admin ./cmd/admin

# Run speech service
SPEECH_DB_DSN="root:pass@tcp(localhost:3306)/octo_speech?parseTime=true" \
VOICE_LITELLM_URL="http://localhost:4000/v1" \
VOICE_LITELLM_KEY="sk-xxx" \
./speech

# Run admin service
SPEECH_DB_DSN="root:pass@tcp(localhost:3306)/octo_speech?parseTime=true" \
ADMIN_USERNAME=admin \
ADMIN_PASSWORD=dev123 \
ADMIN_SECURE_COOKIE=false \
./admin
```

## License

MIT
