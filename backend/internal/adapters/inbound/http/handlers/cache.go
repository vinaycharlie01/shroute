package handlers

import (
	"context"
	"net/http"
	"strconv"

	domaincache "github.com/vinaycharlie01/shroute/backend/internal/domain/cache"
)

// cacheService is the subset of the cache application service the handler
// depends on, declared locally so this package stays decoupled from the
// concrete application/cache import.
type cacheService interface {
	Stats(ctx context.Context) (domaincache.Stats, error)
	List(ctx context.Context, prefix string, limit int) ([]domaincache.Entry, error)
	Flush(ctx context.Context, prefix string) error
}

// Cache handles GET /api/cache/stats, GET /api/cache/entries, and
// POST /api/cache/flush.
type Cache struct {
	svc cacheService
}

// NewCache builds a Cache handler over the given service.
func NewCache(svc cacheService) *Cache {
	return &Cache{svc: svc}
}

// RegisterRoutes implements httpadapter.RouteRegistrar.
func (h *Cache) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/cache/stats", h.Stats)
	mux.HandleFunc("GET /api/cache/entries", h.List)
	mux.HandleFunc("POST /api/cache/flush", h.Flush)
}

type cacheStatsResponse struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	SizeBytes  int64 `json:"size_bytes"`
	EntryCount int64 `json:"entry_count"`
}

type cacheEntryResponse struct {
	Key       string `json:"key"`
	SizeBytes int64  `json:"size_bytes"`
	TTL       string `json:"ttl,omitempty"`
}

type cacheListResponse struct {
	Entries []cacheEntryResponse `json:"entries"`
}

type cacheFlushResponse struct {
	Message string `json:"message"`
}

// Stats handles GET /api/cache/stats.
//
// @Summary      Get cache statistics
// @Description  Returns aggregate cache hit/miss counters, memory usage, and entry count.
// @Tags         Cache
// @Produce      json
// @Success      200  {object}  cacheStatsResponse
// @Failure      500  {object}  errorResponse
// @Router       /api/cache/stats [get].
func (h *Cache) Stats(w http.ResponseWriter, r *http.Request) {
	st, err := h.svc.Stats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})

		return
	}

	writeJSON(w, http.StatusOK, cacheStatsResponse{
		Hits:       st.Hits,
		Misses:     st.Misses,
		SizeBytes:  st.SizeBytes,
		EntryCount: st.EntryCount,
	})
}

// List handles GET /api/cache/entries?prefix=&limit=.
//
// @Summary      List cache entries
// @Description  Returns cache entries matching the given key prefix, up to the requested limit.
// @Tags         Cache
// @Produce      json
// @Param        prefix  query  string  true   "Key prefix to filter by"
// @Param        limit   query  int     false  "Maximum entries to return (default 100)"
// @Success      200     {object}  cacheListResponse
// @Failure      400     {object}  errorResponse
// @Router       /api/cache/entries [get].
func (h *Cache) List(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 100
	}

	entries, err := h.svc.List(r.Context(), prefix, limit)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})

		return
	}

	resp := cacheListResponse{Entries: make([]cacheEntryResponse, len(entries))}
	for i, e := range entries {
		resp.Entries[i] = cacheEntryResponse{
			Key:       e.Key,
			SizeBytes: e.SizeBytes,
			TTL:       e.TTL.String(),
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// Flush handles POST /api/cache/flush{?prefix=}{&all=true}.
// When all=true and no prefix is given it flushes the entire cache.
//
// @Summary      Flush cache entries
// @Description  Removes cache entries matching the given prefix. Pass all=true with no prefix to flush the entire Redis database.
// @Tags         Cache
// @Produce      json
// @Param        prefix  query  string  false  "Key prefix to flush"
// @Param        all     query  bool    false  "When true and no prefix, flushes entire cache"
// @Success      200     {object}  cacheFlushResponse
// @Failure      400     {object}  errorResponse
// @Router       /api/cache/flush [post].
func (h *Cache) Flush(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	all := r.URL.Query().Get("all") == "true"

	// When all=true, pass the empty prefix through so the underlying adapter
	// can issue FLUSHDB. Otherwise the service layer rejects empty prefixes
	// with ErrNoPrefix to guard against accidental full flushes.
	if all && prefix == "" {
		if err := h.svc.Flush(r.Context(), ""); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})

			return
		}

		writeJSON(w, http.StatusOK, cacheFlushResponse{Message: "all cache flushed"})

		return
	}

	if err := h.svc.Flush(r.Context(), prefix); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})

		return
	}

	writeJSON(w, http.StatusOK, cacheFlushResponse{Message: "cache flushed"})
}
