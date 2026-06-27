package http

import (
	"log/slog"
	nethttp "net/http"

	"monomicro-server/internal/rates"
)

type Handlers struct {
	ratesService *rates.Service
	logger       *slog.Logger
}

func NewHandlers(ratesService *rates.Service, logger *slog.Logger) *Handlers {
	return &Handlers{
		ratesService: ratesService,
		logger:       logger,
	}
}

func (h *Handlers) Routes() nethttp.Handler {
	mux := nethttp.NewServeMux()
	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/rates.txt", h.ratesText)

	return h.logRequests(mux)
}

func (h *Handlers) health(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		w.Header().Set("Allow", nethttp.MethodGet)
		nethttp.Error(w, "method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok"))
}

func (h *Handlers) ratesText(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		w.Header().Set("Allow", nethttp.MethodGet)
		nethttp.Error(w, "method not allowed", nethttp.StatusMethodNotAllowed)
		return
	}

	text, err := h.ratesService.RatesText(r.Context())
	if err != nil {
		h.logger.Error("rates request failed", "error", err)
		nethttp.Error(w, "rates unavailable", nethttp.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=60")
	_, _ = w.Write([]byte(text))
}

func (h *Handlers) logRequests(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		next.ServeHTTP(w, r)
		h.logger.Info("http request", "method", r.Method, "path", r.URL.Path, "remote_addr", r.RemoteAddr)
	})
}
