# Task 11: Batches, Files & Storage

**Complexity**: High — first slice needing a genuinely new outbound
dependency (blob/object storage), plus async batch-job lifecycle
(queued → running → completed/failed), not just request/response CRUD.

**TS source**: `OmniRoute/vinaydoc/SLICE_14_BATCHES_FILES_STORAGE.md` —
`/api/batches/*`, `/api/v1/batches/*`, `/api/files/*`, `/api/v1/files/*`,
`/api/storage/health`.

## End-to-end flow

1. **Domain** — `internal/domain/file/file.go`: `File{ID, Name, MimeType
   string, SizeBytes int64, StorageKey string, CreatedAt time.Time}`.
   `internal/domain/batch/batch.go`: `Batch{ID, Status BatchStatus,
   FileIDs []string, ResultFileID *string, CreatedAt, CompletedAt
   *time.Time}`, `BatchStatus` enum (`Queued`/`Running`/`Completed`/`Failed`).
2. **Ports** — `FileRepository` (metadata CRUD), `BlobStore`
   (`Put/Get/Delete(ctx, key, ...)` — the actual byte storage, kept
   separate from metadata exactly like `ProviderRepository`/`ProviderProbe`
   in Task 07), `BatchRepository` (CRUD + status transitions),
   `BatchRunner` (`Enqueue(ctx, batch) error` — kicks off async
   processing) in `ports.go`.
3. **Application** — `internal/application/file/service.go`: validates MIME
   type/size limits before calling `BlobStore.Put`, persists metadata via
   `FileRepository` only after the blob write succeeds (write-blob-then-
   record-metadata ordering, so metadata never points at a missing blob).
   `internal/application/batch/service.go`: `Create` validates referenced
   `FileIDs` exist, enqueues via `BatchRunner`; `Status`/`Result` are reads.
4. **Outbound adapters** — `internal/adapters/outbound/mongodb/{file,batch}.go`
   for metadata; new `internal/adapters/outbound/blobstore/` package
   implementing `BlobStore` — start with a filesystem-backed implementation
   under `DATA_DIR` (mirrors the TS `DATA_DIR` convention) since that
   requires no new infra, with the interface designed so an S3-compatible
   implementation can be swapped in later without touching the application
   layer. `BatchRunner` can start as an in-process goroutine worker pool
   (`internal/adapters/outbound/batchworker/`) reading queued batches from
   Mongo — no new external broker required for v1.
5. **Inbound handler** — `internal/adapters/inbound/http/handlers/{file,batch}.go`:
   multipart upload for `POST /api/files`, `GET /api/files/{id}`,
   `POST /api/batches`, `GET /api/batches/{id}`, `GET /api/storage/health`
   (delegates to `BlobStore` health, registered as a `ports.Pinger` so it
   shows up in `/readyz` for free — reuse Task 01's health aggregation).
6. **Router/DI** — usual extension pattern; `BlobStore`'s storage root comes
   from a new `config.Storage` section.
7. **Tests** — unit tests for upload validation and batch status
   transitions; integration test writing/reading a real file through the
   filesystem `BlobStore`, and a batch end-to-end through the worker pool.

## Checklist

- [ ] `internal/domain/file`, `internal/domain/batch`
- [ ] `FileRepository`, `BlobStore`, `BatchRepository`, `BatchRunner` ports
- [ ] Application services + unit tests
- [ ] Mongo metadata adapters + filesystem `BlobStore` + `batchworker` pool + integration tests
- [ ] Handlers (incl. multipart upload) + router wiring
- [ ] `config.Storage` section + DI wiring; `BlobStore` registered as `ports.Pinger`
- [ ] Full gate: build/vet/fmt/lint/test
