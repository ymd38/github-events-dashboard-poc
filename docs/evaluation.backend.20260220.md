# Software Evaluation Report: backend/ - 2026/02/20

---

## 1. Executive Summary

The Go backend is a well-structured, pragmatic implementation for a real-time webhook dashboard. The layered architecture (handler → service → repository) is sound, the HMAC webhook verification and AES-256-GCM token encryption are correctly implemented, and structured JSON logging is uniformly applied. For a proof-of-concept, this is a solid foundation.

The single largest bottleneck preventing world-class quality is **the complete absence of automated tests**. Every business-logic invariant, security contract, and edge case is unverified. Combined with several architectural shortcuts—a package-level mutable `BroadcastFunc` variable that introduces hidden coupling, concrete-type dependencies that prevent mocking, and a missing request-body size limit on the public webhook endpoint—the codebase carries meaningful risk in production. Addressing testability, the global variable, and the body-size guard are the highest-leverage improvements available.

---

## 2. Multi-Dimensional Scorecard (1-10)

| Category | Score | Strategic Insights |
| :--- | :--- | :--- |
| Principles (SOLID / Idempotency) | 6 / 10 | Good idempotency via `ON DUPLICATE KEY`; broken by the global `BroadcastFunc` (DIP violation) and concrete-type coupling across all layers (OCP/DIP) |
| Resiliency & Reliability | 5 / 10 | Graceful shutdown, connection pool, and SSE backpressure are present; unbounded webhook body reads, blocking `sseHub.Broadcast` call-path, and missing DB query timeouts are significant gaps |
| Observability | 6 / 10 | Uniform structured JSON logging is a genuine strength; no request-ID correlation, no metrics endpoint, no log-level runtime control weakens the overall picture |
| Security | 7 / 10 | HMAC verification, AES-256-GCM, HttpOnly/SameSite cookies, and CSRF state cookie are correctly implemented; `Access-Control-Allow-Origin: *` in SSE handler, missing body-size limit, and weak config validation erode the score |
| DX / Cognitive Load | 6 / 10 | Package structure is clear and idiomatic; duplicated `ptrString` helper, duplicated `sessionName` constant across packages, and zero test coverage impose cognitive cost on maintainers |

**Overall weighted average: 6.0 / 10**

---

## 3. Deep Dive Analysis

### Strategy & Patterns

**Framework & Layering**

The `cmd/server/main.go` wires all dependencies explicitly using constructor injection, which is idiomatic Go and correct. The chi router, connection pool configuration in `db.go`, and graceful 10-second shutdown window are all appropriate choices.

**Dependency Inversion Violation — Package-Level Global Variable**

`backend/internal/handler/webhook.go`, lines 107-113:

```go
// Broadcast callback type for SSE integration (set later in Phase 5).
var BroadcastFunc func(event model.Event)

func broadcastEvent(event model.Event) {
    if BroadcastFunc != nil {
        BroadcastFunc(event)
    }
}
```

This is the most architecturally damaging pattern in the codebase. A mutable package-level variable violates the Dependency Inversion Principle, introduces a hidden dependency that cannot be seen from `WebhookHandler`'s constructor, and makes concurrent testing unsafe without a mutex. The comment "set later in Phase 5" reveals this was a temporary scaffold that became permanent. `WebhookHandler` should accept a `Broadcaster` interface at construction time.

**Concrete-Type Dependencies Block Testability**

Throughout the codebase, handlers and the service depend on concrete struct types, not interfaces:

- `handler/webhook.go` line 22: `eventService *service.EventService`
- `handler/events.go` line 22: `eventService *service.EventService`
- `service/event_service.go` line 17: `repo *repository.EventRepository`
- `auth/oauth.go` line 25: `userRepo *repository.UserRepository`

Because no interface is declared for the service or repository layer, it is impossible to substitute a mock in unit tests without a real database. This is the root cause of zero test coverage.

**Duplicated `ptrString` Helper**

`auth/oauth.go` line 185 and `service/event_service.go` line 168 both define an identical `ptrString(s string) *string` function. This violates DRY and will diverge silently over time. A shared `internal/util` or `internal/ptr` package would eliminate this.

**Duplicated Session Constant**

`auth/session.go` line 11 and `middleware/auth.go` line 15 both declare `sessionName = "github-dashboard-session"` and `sessionUser = "user_id"`. If the session cookie name is ever changed in one location but not the other, authentication will silently break.

**Pagination Edge Case**

`service/event_service.go` line 59:

```go
totalPages := (total + perPage - 1) / perPage
```

When `total` is 0, this correctly returns 0. However when `perPage` is 0 (pathologically), this panics with a division by zero. The handler validates `perPage >= 1`, which prevents the panic via the public API, but the service itself has no guard, violating the Robustness Principle at the layer boundary.

**`OccurredAt` Timestamp Inaccuracy**

`service/event_service.go` lines 121-122 and 162-163:

```go
OccurredAt: time.Now().UTC(),
ReceivedAt: time.Now().UTC(),
```

`OccurredAt` is intended to represent when the GitHub event actually occurred. The correct value should come from the webhook payload (the `created_at` field in the `issue` or `pull_request` object). Using `time.Now()` means `OccurredAt` and `ReceivedAt` are always identical, rendering `OccurredAt` semantically meaningless.

**`truncateString` Corrupts Multi-Byte UTF-8**

`service/event_service.go` lines 175-180:

```go
func truncateString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen]
}
```

`len(s)` in Go returns the byte count, not the Unicode code-point count. Slicing at a byte boundary can split a multi-byte UTF-8 rune, producing an invalid UTF-8 string. Since this value is written to MySQL (utf8mb4) and returned as JSON, the result can be a malformed JSON response or a database error for content containing Japanese, Arabic, emoji, or other multi-byte characters.

**`eventType` Filter is Not Validated**

`handler/events.go` line 40:

```go
eventType := r.URL.Query().Get("event_type")
```

This value is passed directly through to the repository `WHERE event_type = ?` clause via a parameterized query (preventing SQL injection), but no validation restricts it to the known set `{"issues", "pull_request"}`. An attacker can pass arbitrary strings, causing unnecessary database work. While not a security flaw, it is a quality gap.

**Route Ordering Conflict**

`cmd/server/main.go` lines 73-75:

```go
r.Get("/api/events", eventsHandler.List)
r.Get("/api/events/{id}", eventsHandler.GetByID)
r.Get("/api/events/stream", sseHandler.ServeHTTP)
```

Chi resolves `/api/events/stream` before `{id}` because static segments take precedence over path parameters in chi's routing. The current ordering happens to work, but the SSE route is registered after the `{id}` route within the same group. If a developer reorders these lines, the stream endpoint will be captured by `GetByID`, which will attempt to parse "stream" as an int64 and return a 400. This is a latent fragility with no documentation comment to warn maintainers.

**Config Validation is Incomplete**

`config/config.go` lines 61-66:

```go
if cfg.MySQLUser == "" || cfg.MySQLPassword == "" || cfg.MySQLDatabase == "" {
    return nil, fmt.Errorf("MYSQL_USER, MYSQL_PASSWORD, and MYSQL_DATABASE are required")
}
if cfg.TokenEncryptionKey == "" {
    return nil, fmt.Errorf("TOKEN_ENCRYPTION_KEY is required (64 hex characters for AES-256)")
}
```

`GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_WEBHOOK_SECRET`, and `SESSION_SECRET` are not validated as non-empty at startup. The server will start successfully with empty values for all of these, leading to:
- Webhook signature verification always failing (empty secret matches no signature correctly)
- OAuth flow silently broken
- Sessions using an empty key (gorilla/sessions with an empty key is technically valid but insecure)

---

### Reliability & Security

**Critical: No Request Body Size Limit on Public Webhook Endpoint**

`handler/webhook.go` line 36:

```go
body, err := io.ReadAll(r.Body)
```

The `/api/webhook` endpoint is public (no auth middleware) and reads the entire request body into memory with no size limit. A single malicious POST of a multi-gigabyte body can exhaust available memory and crash the server. `http.MaxBytesReader` should wrap `r.Body` before `io.ReadAll`. This is a P0 Denial-of-Service vulnerability.

**`http.DefaultClient` Has No Timeout**

`auth/oauth.go` line 162:

```go
resp, err := http.DefaultClient.Do(req)
```

`http.DefaultClient` has no timeout. If the GitHub API becomes slow or unresponsive, this call will block the goroutine serving the OAuth callback indefinitely. Since Go's HTTP server has a `WriteTimeout` of 15 seconds, this will cause a timeout response to the user, but the goroutine continues to leak until the underlying TCP connection times out (potentially hours). A dedicated `http.Client` with a 10-second timeout should be constructed and stored on `OAuthHandler`.

**SSE Handler Sets `Access-Control-Allow-Origin: *`**

`handler/sse.go` line 31:

```go
w.Header().Set("Access-Control-Allow-Origin", "*")
```

The global CORS middleware correctly sets `Access-Control-Allow-Origin` to the configured `FrontendURL` and sets `Access-Control-Allow-Credentials: true`. The SSE handler then overrides this header with `*`. A CORS header value of `*` cannot be combined with `credentials: 'include'` in browsers (the browser will block the response). This creates an inconsistency that either breaks the SSE connection for credentialed requests or, if the frontend works around it by not sending credentials, creates a security gap by exposing the SSE stream to any origin. The SSE handler should not set this header independently; the global middleware is sufficient.

**`sseHub.Broadcast` Can Block Under Load**

`sse/hub.go` line 26: the broadcast channel has a buffer of 256. `cmd/server/main.go` line 49-51 sets `BroadcastFunc` to call `sseHub.Broadcast` synchronously. Under heavy webhook ingestion or with 256+ slow clients, the broadcast channel can fill, causing the webhook handler goroutine to block waiting for capacity. This means slow SSE clients can exert backpressure that delays webhook processing and acknowledgment to GitHub.

**Double-Locking Race Condition in SSE Hub**

`sse/hub.go` lines 59-68:

```go
h.mu.RLock()
for client := range h.clients {
    select {
    case client <- data:
    default:
        go func(c chan []byte) {
            h.unregister <- c
        }(client)
    }
}
h.mu.RUnlock()
```

While holding `mu.RLock`, the code spawns goroutines that send to `h.unregister`. The `unregister` case in the `Run()` loop (`hub.go` line 41) attempts to acquire `mu.Lock()`. Since `Run()` is a single goroutine and is currently blocked in the `broadcast` case, the goroutine writing to `h.unregister` will block until `Run()` loops back around. However, the current `mu.RLock()` is held by the broadcast case. The `Run()` goroutine cannot re-enter to process `unregister` while the broadcast case is still executing. This creates a potential deadlock when the channel buffer fills: the unregister goroutine blocks on `h.unregister <- c` because `Run()` is blocked inside the broadcast case still holding `RLock`, and `Run()` cannot process the unregister message until the broadcast case completes. With default sends the client channel is full, so this path is triggered precisely when the system is under load.

**DB Query Context / Timeout Missing**

All repository queries (`event_repository.go`, `user_repository.go`) use `r.db.Exec`, `r.db.Query`, and `r.db.QueryRow` without a context. This means a slow or hung MySQL query cannot be cancelled when the HTTP client disconnects, and queries have no independent deadline. This can exhaust the connection pool under load. All DB calls should accept and propagate a `context.Context` from the HTTP request.

**`io.ReadAll` After `defer r.Body.Close()`**

`handler/webhook.go` lines 36-42:

```go
body, err := io.ReadAll(r.Body)
if err != nil { ... }
defer r.Body.Close()
```

`io.ReadAll` is called before `defer r.Body.Close()`. While this works correctly in Go (the body is read before the function returns and Close is deferred), convention is to `defer r.Body.Close()` immediately after receiving the request, before any reads. The current ordering is functionally correct but subtly non-idiomatic; if the order of lines was changed during refactoring, a bug could be introduced.

**`Me` Endpoint Has No Auth Middleware**

`cmd/server/main.go` line 69:

```go
r.Get("/api/auth/me", oauthHandler.Me)
```

`/api/auth/me` is registered outside the protected group and does its own manual session check. This is a consistent implementation (the handler correctly returns 401 without a session), but it means the auth middleware is not applied, and a future developer might add additional logic to `Me` without realising it runs without the middleware's protections (such as context injection of `userID`).

**`sessionName` Duplication Creates Security Drift Risk**

As noted above, `auth/session.go:11` and `middleware/auth.go:15` both define the string `"github-dashboard-session"`. Two independent `sessions.CookieStore` instances are also created: one in `auth/session.go:22` and one in `cmd/server/main.go:54`. Only one of these stores is passed to the Auth middleware. The Auth middleware store and the SessionManager store are both initialized from the same `cfg.SessionSecret`, so they will decode each other's cookies correctly — but this is a coincidence of shared configuration, not an enforced invariant. If the SessionManager is ever given a different secret (for key rotation), the middleware will start rejecting all sessions silently.

---

### Observability & Operability

**No Request-ID / Correlation ID**

The logger middleware (`middleware/logger.go`) does not generate or propagate a request ID. All log events from a single request lifecycle (webhook received → event saved → SSE broadcast) cannot be correlated in a log aggregator without a shared trace ID. In production, this makes debugging specific failures very difficult.

**No Metrics Endpoint**

There is no `/metrics` endpoint (Prometheus or otherwise). Operational visibility into request rates, error rates, DB pool utilization, SSE client count, and event processing latency requires log scraping, which is fragile and slow.

**Log Level Cannot Be Changed at Runtime**

`middleware/logger.go` line 11:

```go
var jsonLogger = log.New(os.Stdout, "", 0)
```

The log level is not configurable. "Debug" vs "Info" vs "Warn" cannot be toggled without a code change and redeployment. In production environments, this prevents dynamic verbosity adjustment for incident investigation.

**`json.Marshal` Error Silently Ignored in Logger**

`middleware/logger.go` line 57:

```go
data, _ := json.Marshal(entry)
```

If marshaling fails (which is practically impossible for this struct but is a code pattern concern), the log line is silently dropped. The same pattern appears in `LogEvent` at line 72. This is acceptable for a logger but is worth noting as a pattern to avoid propagating to business logic.

---

## 4. Strategic Improvement Roadmap

### P0: Immediate Stabilization (Security & Reliability Blockers)

These issues present active risk to availability or security in any environment.

1. **Add `http.MaxBytesReader` to the webhook endpoint.**
   File: `backend/internal/handler/webhook.go` line 36.
   Wrap `r.Body` with `http.MaxBytesReader(w, r.Body, 1<<20)` (1 MB limit) before calling `io.ReadAll`. This closes the DoS attack surface on the only unauthenticated write endpoint.

2. **Fix the `Access-Control-Allow-Origin: *` override in the SSE handler.**
   File: `backend/internal/handler/sse.go` line 31.
   Remove the manual `Access-Control-Allow-Origin` header set. The global CORS middleware already sets the correct origin. The current override breaks credentialed cross-origin SSE and may expose the stream to unintended origins.

3. **Add a timeout to the GitHub API HTTP client.**
   File: `backend/internal/auth/oauth.go` line 162.
   Replace `http.DefaultClient` with a locally scoped `&http.Client{Timeout: 10 * time.Second}` stored as a field on `OAuthHandler`. This prevents goroutine leaks and unresponsive OAuth callbacks.

4. **Fix `truncateString` to be UTF-8 safe.**
   File: `backend/internal/service/event_service.go` lines 175-180.
   Replace the byte-slice truncation with `[]rune(s)` conversion before slicing, or use `utf8.RuneCountInString` and iterate with `utf8.DecodeRuneInString`. This prevents garbled data being written to MySQL and returned in API responses.

5. **Validate all required secrets at startup.**
   File: `backend/internal/config/config.go`.
   Add non-empty validation for `GITHUB_CLIENT_ID`, `GITHUB_CLIENT_SECRET`, `GITHUB_WEBHOOK_SECRET`, and `SESSION_SECRET`. A server that starts with empty secrets gives a false sense of readiness.

### P1: Architectural Debt (Correctness & Testability)

These issues block the ability to write meaningful tests and introduce latent correctness bugs.

6. **Replace `BroadcastFunc` global with a constructor-injected `Broadcaster` interface.**
   File: `backend/internal/handler/webhook.go` lines 107-113.
   Define `type Broadcaster interface { Broadcast(model.Event) }`. Pass the `sseHub` as this interface to `NewWebhookHandler`. Remove the package-level variable entirely. This enables mocking in tests and eliminates the hidden coupling.

7. **Introduce repository and service interfaces.**
   Define `EventRepository` and `UserRepository` interfaces in `internal/repository/` (or a separate `internal/port/` package). Have all handlers and services depend on the interface, not the concrete struct. This is the prerequisite for all unit testing.

8. **Write unit tests for all business-critical paths.**
   Priority order:
   - `crypto.go`: Encrypt/Decrypt round-trip, short key rejection, tampered ciphertext detection.
   - `service/event_service.go`: `processIssueEvent` (action != "opened" returns nil), `processPullRequestEvent` (not merged returns nil, merged returns event), `truncateString` with multi-byte input.
   - `handler/webhook.go`: Valid HMAC passes, invalid HMAC returns 401, missing headers return 400, duplicate delivery returns 200 with "duplicate" status.
   - `config/config.go`: Missing required variables produce errors.

9. **Fix `OccurredAt` to use the payload timestamp.**
   File: `backend/internal/service/event_service.go` lines 121 and 162.
   Add `CreatedAt string json:"created_at"` to `issuePayload.Issue` and `pullRequestPayload.PullRequest`. Parse the timestamp with `time.Parse(time.RFC3339, p.Issue.CreatedAt)` and set `OccurredAt` from the payload. Fall back to `time.Now()` if the field is empty.

10. **Eliminate the duplicated `ptrString` function.**
    Create `backend/internal/ptr/ptr.go` exporting `func String(s string) *string`. Replace both usages in `service/event_service.go` and `auth/oauth.go`.

11. **Eliminate the duplicated `sessionName` / `sessionStore` instances.**
    File: `cmd/server/main.go` line 52-60 and `auth/session.go` line 22.
    The `sessions.CookieStore` created at line 54 of `main.go` should be the single source of truth. Pass it into `NewSessionManager` rather than letting `SessionManager` create its own store from the same secret.

12. **Propagate `context.Context` through all DB calls.**
    Files: `repository/event_repository.go` and `repository/user_repository.go`.
    Replace `db.Exec`, `db.Query`, and `db.QueryRow` with their `Context` variants (`ExecContext`, `QueryContext`, `QueryRowContext`). Pass the HTTP request context from handlers through the service layer to the repository. This enables query cancellation and timeout control.

### P2: Technical Debt (Scalability & Observability)

13. **Add a request-correlation-ID to the logger middleware.**
    File: `backend/internal/middleware/logger.go`.
    Generate a UUID (or read `X-Request-ID` from an incoming header) at the start of each request. Store it in the request context and include it in all `LogEvent` calls for that request. This makes log correlation in production feasible.

14. **Validate the `event_type` query parameter.**
    File: `backend/internal/handler/events.go` line 40.
    Restrict valid values to `""`, `"issues"`, and `"pull_request"`. Return a 400 for any other value. This prevents unnecessary DB queries and produces a better developer experience for API consumers.

15. **Add a `perPage` guard in the service layer.**
    File: `backend/internal/service/event_service.go` line 59.
    Add `if perPage <= 0 { return nil, fmt.Errorf("perPage must be positive") }` at the top of `ListEvents`. Do not rely solely on the handler layer to enforce this invariant.

16. **Document the route ordering dependency.**
    File: `cmd/server/main.go` lines 73-75.
    Add a comment above the SSE route noting that `/api/events/stream` must be registered before `{id}` to avoid the static segment being captured by the wildcard parameter. This prevents a class of latent refactoring bugs.

### P3: Excellence (World-Class Observability & Operability)

17. **Add a Prometheus `/metrics` endpoint.**
    Instrument: request count/latency by route and status code, active SSE client count (`hub.ClientCount()` is already available), DB connection pool stats, event processing count by type. A single `go.opentelemetry.io/otel` or `github.com/prometheus/client_golang` integration provides significant operational value.

18. **Add an environment-controlled log level.**
    Read `LOG_LEVEL` from the environment in `config.go`. Pass it to the logger. Suppress debug-level `LogEvent` calls below the configured threshold.

19. **Consider context propagation for SSE client count logging.**
    File: `sse/hub.go` lines 38-39.
    `ClientCount()` acquires `mu.RLock()` while the hub is already holding `mu.Lock()` (inside the register case). While the lock is `sync.RWMutex` and multiple readers are allowed, calling `ClientCount()` from within the `register` case (which holds the write lock) means `ClientCount()` tries to acquire a read lock while the write lock is held by the same goroutine. In Go, `sync.RWMutex` does not support recursive locking — a goroutine holding a write lock cannot also acquire a read lock. This will **deadlock** on the `register` and `unregister` paths. The client count should be read directly from `len(h.clients)` inside the lock-holding cases, not via the `ClientCount()` method.

---

## Appendix: Evidence Index

| Finding | File | Line(s) |
| :--- | :--- | :--- |
| Global `BroadcastFunc` (DIP violation) | `handler/webhook.go` | 107-113 |
| Concrete service dependency in handler | `handler/webhook.go` | 22 |
| Concrete service dependency in handler | `handler/events.go` | 22 |
| Concrete repo dependency in service | `service/event_service.go` | 17 |
| `truncateString` byte-slices UTF-8 strings | `service/event_service.go` | 175-180 |
| `OccurredAt` uses `time.Now()` not payload | `service/event_service.go` | 121, 162 |
| Duplicated `ptrString` | `service/event_service.go` L168 / `auth/oauth.go` L185 | 168 / 185 |
| Unbounded `io.ReadAll` on public endpoint | `handler/webhook.go` | 36 |
| `http.DefaultClient` has no timeout | `auth/oauth.go` | 162 |
| SSE overrides CORS to `*` | `handler/sse.go` | 31 |
| Duplicated `sessionName` constant | `auth/session.go` L11 / `middleware/auth.go` L15 | 11 / 15 |
| Two `CookieStore` instances | `auth/session.go` L22 / `cmd/server/main.go` L54 | 22 / 54 |
| No validation for OAuth/webhook secrets | `config/config.go` | 61-66 |
| No context propagation to DB layer | `repository/event_repository.go` | all methods |
| `ClientCount()` deadlock inside hub lock | `sse/hub.go` | 38-39, 49-50 |
| Route ordering fragility | `cmd/server/main.go` | 73-75 |
| No request-correlation-ID | `middleware/logger.go` | entire file |
| Zero test files | entire backend | — |
