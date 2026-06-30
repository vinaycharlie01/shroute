# 🎯 Slice 14: Go Backend for Batches, Files & Storage Routes

**Goal**: Migrate batch processing, file upload/download, and storage health endpoints from TypeScript to Go.

**Routes**: `/api/batches/*`, `/api/v1/batches/*`, `/api/files/*`, `/api/v1/files/*`, `/api/storage/health`

---

## ✅ TASK 1: Types

**Files to create**: `pkg/types/batch.go`, `pkg/types/file.go`

```go
type Batch struct {
    ID            string `json:"id"`
    Status        string `json:"status"` // "validating", "in_progress", "completed", "failed", "cancelled"
    Endpoint      string `json:"endpoint"`
    InputFileID   string `json:"input_file_id"`
    OutputFileID  string `json:"output_file_id,omitempty"`
    ErrorFileID   string `json:"error_file_id,omitempty"`
    CreatedAt     string `json:"created_at"`
    CompletedAt   string `json:"completed_at,omitempty"`
    RequestCount  int    `json:"request_count"`
    CompletedCount int   `json:"completed_count"`
    FailedCount   int    `json:"failed_count"`
}

type UploadedFile struct {
    ID          string `json:"id"`
    Bytes       int64  `json:"bytes"`
    Filename    string `json:"filename"`
    Purpose     string `json:"purpose"` // "batch", "fine-tune", "assistants"
    CreatedAt   string `json:"created_at"`
    ContentType string `json:"content_type"`
}
```

| # | Step | Done |
|---|------|------|
| 1.1 | Create types | ☐ |
| 1.2 | Run `go build` | ☐ |

---

## ✅ TASK 2: Repositories

**Files**: `internal/db/batch.go`, `internal/db/file.go`

```go
type BatchRepo struct { db *sql.DB }
func (r *BatchRepo) List() ([]types.Batch, error)
func (r *BatchRepo) Get(id string) (*types.Batch, error)
func (r *BatchRepo) Create(batch *types.Batch) error
func (r *BatchRepo) Cancel(id string) error
func (r *BatchRepo) DeleteCompleted() error
func (r *BatchRepo) UpdateStatus(id, status string) error

type FileRepo struct { db *sql.DB }
func (r *FileRepo) List() ([]types.UploadedFile, error)
func (r *FileRepo) Get(id string) (*types.UploadedFile, error)
func (r *FileRepo) Upload(file *types.UploadedFile, content []byte) error
func (r *FileRepo) GetContent(id string) ([]byte, error)
func (r *FileRepo) Delete(id string) error
```

| # | Step | Done |
|---|------|------|
| 2.1 | Implement batch CRUD | ☐ |
| 2.2 | Implement file upload/storage | ☐ |
| 2.3 | Write tests | ☐ |

---

## ✅ TASK 3: Handlers

**Files**: `api/handlers/batch.go`, `api/handlers/file.go`

| Route | Handler | Done |
|-------|---------|------|
| `GET /api/batches` | `ListBatches` | ☐ |
| `POST /api/batches` | `CreateBatch` | ☐ |
| `GET /api/batches/:id` | `GetBatch` | ☐ |
| `POST /api/batches/:id/cancel` | `CancelBatch` | ☐ |
| `POST /api/batches/delete-completed` | `DeleteCompletedBatches` | ☐ |
| `GET /api/files` | `ListFiles` | ☐ |
| `POST /api/files` | `UploadFile` | ☐ |
| `GET /api/files/:id` | `GetFile` | ☐ |
| `DELETE /api/files/:id` | `DeleteFile` | ☐ |
| `GET /api/files/:id/content` | `GetFileContent` | ☐ |
| `GET /api/storage/health` | `StorageHealthCheck` | ☐ |

| # | Step | Done |
|---|------|------|
| 3.1 | Wire all routes | ☐ |
| 3.2 | Test file upload → list → download | ☐ |
| 3.3 | Test batch create → cancel → list | ☐ |

---

## ✅ TASK 4: Sidecar + Tests

| # | Step | Done |
|---|------|------|
| 4.1 | Update nginx | ☐ |
| 4.2 | `go test ./...` | ☐ |
| 4.3 | Deploy | ☐ |

---

## 🚀 QUICK START

```bash
cd omniroute-go && go run .
curl localhost:8080/api/files
curl localhost:8080/api/batches
curl localhost:8080/api/storage/health