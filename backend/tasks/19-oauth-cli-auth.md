# Task 19: OAuth & CLI Auth

**Complexity**: Very high and security-critical — multi-provider OAuth
flows (Claude Code, Codex, Cursor, Kiro, Zed, AGY) with token
storage/refresh, plus credential import/export. Mistakes here have the
highest blast radius of any CRUD-shaped slice (leaked/forged credentials),
so build it after the simpler auth groundwork in Task 04 (API keys, the
`Hasher`/`crypto` adapter) is proven out.

**TS source**: `OmniRoute/vinaydoc/SLICE_11_OAUTH_CLI_AUTH.md` —
`/api/oauth/*`, `/api/providers/{id}/claude-auth/*`,
`/api/providers/{id}/codex-auth/*`, `/api/providers/agy-auth/*`,
`/api/providers/zed/*`, `/api/cloud/auth`, `/api/cloud/credentials/update`.
Cross-reference `src/lib/oauth/` and `docs/security/PUBLIC_CREDS.md` in the
OmniRoute TS repo — **CLAUDE.md hard rule #11 carries over directly**: any
public upstream OAuth client_id/secret or Firebase Web key extracted from a
public CLI must be embedded via an equivalent of `resolvePublicCred()`,
never as a Go string literal.

## End-to-end flow

1. **Domain** — `internal/domain/oauth/oauth.go`: `Token{ProviderID,
   AccessToken, RefreshToken []byte (encrypted), ExpiresAt time.Time}`,
   `Flow{ProviderID, State, CodeVerifier string}` (PKCE state, short-lived).
2. **Ports** — `OAuthTokenRepository` (CRUD, values always passed through
   already-encrypted), `OAuthProvider` (`AuthURL(ctx, state string)
   string`, `Exchange(ctx, code string) (Token, error)`,
   `Refresh(ctx, t Token) (Token, error)` — one implementation per upstream
   CLI's OAuth dialect: Claude, Codex, Cursor, Kiro, Zed, AGY), reuse the
   `Hasher`/encryption adapter pattern from Task 04 for token-at-rest
   encryption (AES-256-GCM, per CLAUDE.md's standing security rule) in
   `ports.go`.
3. **Application** — `internal/application/oauth/service.go`: `StartFlow`
   generates PKCE state, persists the short-lived `Flow`; `Complete`
   validates the returned `state` matches, calls the matching
   `OAuthProvider.Exchange`, encrypts and persists the resulting `Token`;
   a background-triggered `RefreshExpiring(ctx)` calls `Refresh` for tokens
   nearing expiry — model this as a port-driven call from a scheduler in
   `cmd/server/main.go` or a small ticker in `di.Container`, not a
   goroutine started inside the application service itself (keep
   application layer free of its own concurrency/scheduling).
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/oauth.go`
   (tokens stored already-encrypted — the Mongo adapter never sees
   plaintext secrets); one `internal/adapters/outbound/oauthcli/{claude,
   codex,cursor,kiro,zed,agy}.go` file per upstream CLI implementing
   `OAuthProvider`, each embedding its public client credentials via the
   `resolvePublicCred()`-equivalent helper (port that helper itself into Go
   as `internal/infrastructure/publiccreds/publiccreds.go` before writing
   any provider file, since every one of these six files depends on it).
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/oauth.go`:
   `GET /api/oauth/{provider}/authorize`, `GET /api/oauth/{provider}/callback`,
   `POST /api/cloud/credentials/update`.
6. **Router/DI** — usual extension pattern; wire all six `OAuthProvider`
   implementations into a small registry in `di.Container` keyed by
   provider ID.
7. **Tests** — unit tests per `OAuthProvider` implementation against
   `httptest.Server` mocking each upstream's token endpoint; a dedicated
   test asserting `publiccreds` values are never hardcoded as raw string
   literals in any `oauthcli/*.go` file (mirrors the TS "publicCreds shape
   assertion" test convention referenced in CLAUDE.md); integration test
   for encrypted token round-trip through Mongo.

## Checklist

- [ ] `internal/infrastructure/publiccreds` (Go port of `resolvePublicCred()`) — build first
- [ ] `internal/domain/oauth`
- [ ] `OAuthTokenRepository`, `OAuthProvider` ports
- [ ] `internal/application/oauth/service.go` (PKCE flow, encrypt-before-persist, refresh) + unit tests
- [ ] Mongo adapter (tokens always encrypted) + 6 `oauthcli/*` provider adapters (via `publiccreds`) + unit tests per provider
- [ ] Test asserting no raw public-cred string literals
- [ ] Handlers + router wiring
- [ ] DI wiring (provider registry + refresh scheduler outside the application layer)
- [ ] Full gate: build/vet/fmt/lint/test
