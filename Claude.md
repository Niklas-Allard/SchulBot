# CLAUDE.md

## Project Overview
Build a production-oriented Go service that connects directly to an IServ mail account via IMAP and SMTP.

The service watches incoming emails, detects command tags in subject or body, executes the matching action, and replies by email to the original sender.

The first supported action is AI prompting via `#ki` or `/ki`.
The architecture must be extensible so additional commands such as `#tasks` and `#calendar` can be added cleanly later.

## Real Mail Environment
Use this real mail setup as the default target environment:

### Preferred SSL/TLS configuration
- IMAP host: `pggv.de`
- IMAP port: `993`
- IMAP security: `SSL/TLS`
- SMTP host: `pggv.de`
- SMTP port: `465`
- SMTP security: `SSL/TLS`
- Authentication: normal password

### Alternative STARTTLS configuration
- IMAP host: `pggv.de`
- IMAP port: `143`
- IMAP security: `STARTTLS`
- SMTP host: `pggv.de`
- SMTP port: `587`
- SMTP security: `STARTTLS`
- Authentication: normal password

Do not hardcode credentials.
Load all secrets from environment variables.

## Core Goal
Create a self-hostable Go daemon that:

1. Connects to the IServ mailbox.
2. Watches for new incoming emails.
3. Detects command tags in subject or body.
4. Routes the request to the correct handler.
5. Sends a reply email with the result.
6. Prevents loops, duplicate processing, and unsafe behavior.

## Command System
Design the system around command tags, not around a single hardcoded AI feature.

### Initial supported tags
- `#ki`
- `/ki`

### Planned or near-future tags
- `#tasks`
- `/tasks`
- `#calendar`
- `/calendar`

### Extensibility requirement
The architecture must make it easy to add further tags later, for example:
- `#translate`
- `#summary`
- `#todo`
- `#search`

Implement a generic dispatcher pattern.

## Trigger Rules
A mail should be treated as a command request when one of the following is true:

- The subject starts with a supported tag
- The first non-empty line of the plain-text body starts with a supported tag

Parsing behavior:
- The first token is the command tag
- The remaining text is the command payload
- For `#ki`, the payload is the AI prompt
- For future commands, the payload should be passed to the relevant handler

## Non-Goals
Do not build:
- an IServ plugin
- a web frontend for v1
- OAuth login UI
- a full task/calendar sync engine in v1
- unnecessary enterprise abstractions

## Architecture
Use a clean modular structure.

Preferred layout:

- `cmd/server/main.go`
- `internal/config/`
- `internal/logging/`
- `internal/mail/`
- `internal/parser/`
- `internal/commands/`
- `internal/handlers/ki/`
- `internal/handlers/tasks/`
- `internal/handlers/calendar/`
- `internal/ai/`
- `internal/app/`
- `internal/store/`
- `internal/model/`
- `internal/testutil/`

## Handler Design
Use a handler interface for commands, for example:

- `CanHandle(tag string) bool`
- `Handle(ctx context.Context, req CommandRequest) (CommandResponse, error)`

`CommandRequest` should contain:
- message metadata
- sender
- subject
- parsed tag
- payload
- raw text if needed

`CommandResponse` should contain:
- reply subject
- reply body
- optional status metadata

## MVP Scope
For v1, implement fully:
- mail watcher
- command parser
- `#ki` handler
- reply sending
- duplicate protection
- loop protection
- configuration
- tests
- Docker support

For `#tasks` and `#calendar`:
- add parser awareness
- add handler stubs or placeholder handlers
- document how those integrations can be implemented next

## AI Integration
Create an AI provider abstraction:

- `Generate(ctx context.Context, prompt string) (string, error)`

Requirements:
- configurable provider
- configurable model
- configurable API URL
- timeout handling
- clean error wrapping

## Future Integrations
Plan for external integrations such as:

### Google Tasks
The system may later support `#tasks` to create or list tasks via Google Tasks API.

### Google Calendar
The system may later support `#calendar` to create, inspect, or summarize events via Google Calendar API.

Important:
- Do not fully implement Google OAuth unless needed for the requested scope
- Prepare the code so such integrations can be added without rewriting the core mail pipeline

## Mail Handling Requirements
- Support IMAP over SSL/TLS and STARTTLS
- Support SMTP over SSL/TLS and STARTTLS
- Parse multipart messages safely
- Prefer `text/plain`
- Reply with proper headers
- Ignore bot self-mails
- Ignore obvious auto-replies when possible
- Log skipped messages with reason

## Reliability and Safety
Implement:
- duplicate detection using message ID or a fallback fingerprint
- self-reply prevention
- empty payload rejection with a helpful answer
- maximum payload size
- maximum response size
- timeouts
- no secret logging

## Storage
For MVP, use a simple persistent processed-message store.
Good options:
- BoltDB
- SQLite
- file-backed state store

Choose the simplest robust option.

## Config
Configuration must come from environment variables.

Include at least:
- `APP_ENV`
- `LOG_LEVEL`

### IMAP
- `IMAP_HOST`
- `IMAP_PORT`
- `IMAP_USERNAME`
- `IMAP_PASSWORD`
- `IMAP_MAILBOX`
- `IMAP_SECURITY`

### SMTP
- `SMTP_HOST`
- `SMTP_PORT`
- `SMTP_USERNAME`
- `SMTP_PASSWORD`
- `SMTP_FROM_NAME`
- `SMTP_FROM_ADDRESS`
- `SMTP_SECURITY`

### AI
- `AI_PROVIDER`
- `AI_API_URL`
- `AI_API_KEY`
- `AI_MODEL`

### Runtime
- `POLL_INTERVAL`
- `MAX_PAYLOAD_CHARS`
- `MAX_RESPONSE_CHARS`

## Coding Principles
- Keep code explicit and maintainable
- Favor simple interfaces
- Avoid overengineering
- Keep files reasonably small
- Use wrapped errors
- Use structured logs
- Write tests for core parsing and dispatch behavior

## Tests
Write tests for:
- tag parsing
- payload extraction
- supported tag routing
- duplicate detection
- self-loop prevention
- reply construction
- config validation

## Deliverables
Claude should create:
- complete Go module
- working MVP for `#ki`
- parser support for `#tasks` and `#calendar`
- placeholder handlers for future commands
- `.env.example`
- `README.md`
- `Dockerfile`
- `docker-compose.yml`
- `Makefile`
- tests

## Workflow Instructions for Claude
When generating code:

1. First show the proposed file tree.
2. Then explain the architecture briefly.
3. Then generate the files in implementation order.
4. Keep the project runnable as early as possible.
5. Prefer a stable polling watcher first; IMAP IDLE may be added if it remains understandable.
6. Use real libraries only.
7. Do not invent external API contracts.
8. For future handlers like `#tasks` and `#calendar`, create minimal placeholder implementations plus clear extension points.

## Definition of Done
The project is done when:
- the app can connect to IServ mail
- a matching `#ki` mail is detected
- the payload is parsed correctly
- the AI backend is called
- a reply mail is sent
- processed mails are not duplicated
- self-reply loops are prevented
- the codebase is ready for future `#tasks` and `#calendar` handlers
- Docker setup and README are complete