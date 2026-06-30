# Task 04: API Keys

**Complexity**: Moderate — CRUD is simple, but keys are security-sensitive:
secrets must be hashed/encrypted at rest (CLAUDE.md hard rule: "Encrypt
credentials at rest (AES-256-GCM)") and scoped (groups/permissions), so this
sits above the pure-CRUD tasks.

**TS source**: `OmniRoute/vinaydoc/SLICE_03_API_KEYS.md` — `/api/keys/*`,
`/api/keys/groups/*`.

## End-to-end flow

1. **Domain** — `internal/domain/apikey/apikey.go`: `Key{ID, Name,
   HashedSecret []byte, GroupID string, Scopes []string, CreatedAt,
   ExpiresAt *time.Time, RevokedAt *time.Time}`, `Group{ID, Name,
   RateLimit int}`. The raw secret is generated once at creation and never
   stored — only its hash, consistent with `extractApiKey`/`isValidApiKey`
   conventions on the TS side.
2. **Ports** — `ApiKeyRepository` (`Create/Get/List/Revoke`),
   `ApiKeyGroupRepository` (`Create/Get/List/Update`) in `ports.go`. Add a
   small `Hasher` port (`Hash([]byte) []byte`, `Verify(hash, secret []byte) bool`)
   so the application layer never imports a concrete crypto library directly.
3. **Application** — `internal/application/apikey/service.go`: generates the
   secret (crypto/rand), hashes it via the `Hasher` port before persisting,
   returns the plaintext secret to the caller exactly once (creation
   response only).
4. **Outbound adapter** — `internal/adapters/outbound/mongodb/apikey.go`:
   `api_keys` collection (unique index on a deterministic lookup hash, not
   the secret itself), `api_key_groups` collection. A `crypto` adapter
   (`internal/adapters/outbound/crypto/hasher.go`, e.g. bcrypt or
   AES-256-GCM-backed) implements the `Hasher` port — first non-Mongo/Redis
   outbound adapter added; this is the intended way to **extend** the
   hexagonal architecture per CLAUDE.md.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/apikey.go`:
   `POST/GET/DELETE /api/keys`, `GET/POST /api/keys/groups`. Auth middleware
   (new `internal/adapters/inbound/http/middleware/auth.go`) validates
   incoming requests against this same `ApiKeyRepository` port — this is the
   slice that should introduce request authentication for the Go backend.
6. **Router/DI** — wire the new auth middleware into the chain in
   `router.go` (after `RequestID`, before route dispatch) and the handler
   per the usual pattern.
7. **Tests** — unit tests for hash/verify round-trip and scope checks;
   integration test creating a key, verifying the secret is never readable
   back in plaintext from Mongo.

## Checklist

- [ ] `internal/domain/apikey`
- [ ] `ApiKeyRepository`, `ApiKeyGroupRepository`, `Hasher` ports
- [ ] `internal/application/apikey/service.go` + unit tests
- [ ] Mongo adapter + new `crypto` outbound adapter + integration test
- [ ] Auth middleware + handlers + router wiring
- [ ] DI wiring
- [ ] Full gate: build/vet/fmt/lint/test
