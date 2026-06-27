package http

import (
	"context"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"monomicro-server/internal/monobank"
	"monomicro-server/internal/rates"
)

type errorProvider struct {
	err error
}

func (p errorProvider) FetchCurrencyRates(_ context.Context) ([]monobank.CurrencyRate, error) {
	return nil, p.err
}

type staticProvider struct {
	rates []monobank.CurrencyRate
}

func (p staticProvider) FetchCurrencyRates(_ context.Context) ([]monobank.CurrencyRate, error) {
	return p.rates, nil
}

func TestRatesTextReturnsGatewayTimeoutForUpstreamTimeout(t *testing.T) {
	service := rates.NewService(errorProvider{err: context.DeadlineExceeded}, 5*time.Minute)
	handler := NewHandlers(service, slog.New(slog.NewTextHandler(testWriter{t: t}, nil))).Routes()

	req := httptest.NewRequest(nethttp.MethodGet, "/rates.txt", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != nethttp.StatusGatewayTimeout {
		t.Fatalf("expected status %d, got %d", nethttp.StatusGatewayTimeout, resp.Code)
	}
}

func TestRatesTextReturnsBadGatewayForInvalidUpstreamRates(t *testing.T) {
	service := rates.NewService(staticProvider{
		rates: []monobank.CurrencyRate{
			{CurrencyCodeA: 840, CurrencyCodeB: 980, RateBuy: 40, RateSell: 41},
		},
	}, 5*time.Minute)
	handler := NewHandlers(service, slog.New(slog.NewTextHandler(testWriter{t: t}, nil))).Routes()

	req := httptest.NewRequest(nethttp.MethodGet, "/rates.txt", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != nethttp.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", nethttp.StatusBadGateway, resp.Code)
	}
}

func TestHealthRejectsUnsupportedMethods(t *testing.T) {
	service := rates.NewService(errorProvider{}, 5*time.Minute)
	handler := NewHandlers(service, slog.New(slog.NewTextHandler(testWriter{t: t}, nil))).Routes()

	req := httptest.NewRequest(nethttp.MethodPost, "/health", nil)
	resp := httptest.NewRecorder()

	handler.ServeHTTP(resp, req)

	if resp.Code != nethttp.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", nethttp.StatusMethodNotAllowed, resp.Code)
	}
}

type testWriter struct {
	t *testing.T
}

func (w testWriter) Write(p []byte) (int, error) {
	w.t.Log(string(p))
	return len(p), nil
}
