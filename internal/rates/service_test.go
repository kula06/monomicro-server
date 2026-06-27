package rates

import (
	"context"
	"errors"
	"testing"
	"time"

	"monomicro-server/internal/monobank"
)

type fakeProvider struct {
	calls int
	rates []monobank.CurrencyRate
	err   error
}

func (p *fakeProvider) FetchCurrencyRates(_ context.Context) ([]monobank.CurrencyRate, error) {
	p.calls++
	return p.rates, p.err
}

func TestParseRatesText(t *testing.T) {
	text, err := ParseRatesText([]monobank.CurrencyRate{
		{CurrencyCodeA: 840, CurrencyCodeB: 980, RateBuy: 40.1, RateSell: 41.2},
		{CurrencyCodeA: 978, CurrencyCodeB: 980, RateBuy: 43, RateSell: 44.25},
		{CurrencyCodeA: 985, CurrencyCodeB: 980, RateBuy: 10.1234567, RateSell: 10.5},
		{CurrencyCodeA: 826, CurrencyCodeB: 980, RateBuy: 50, RateSell: 51},
	})
	if err != nil {
		t.Fatalf("ParseRatesText returned error: %v", err)
	}

	want := "USD|40.1|41.2\nEUR|43|44.25\nPLN|10.123457|10.5\n"
	if text != want {
		t.Fatalf("unexpected rates text\nwant: %q\n got: %q", want, text)
	}
}

func TestParseRatesTextUsesRateCrossWhenBuySellMissing(t *testing.T) {
	text, err := ParseRatesText([]monobank.CurrencyRate{
		{CurrencyCodeA: 840, CurrencyCodeB: 980, RateCross: 40},
		{CurrencyCodeA: 978, CurrencyCodeB: 980, RateCross: 43},
		{CurrencyCodeA: 985, CurrencyCodeB: 980, RateCross: 10},
	})
	if err != nil {
		t.Fatalf("ParseRatesText returned error: %v", err)
	}

	want := "USD|40|40\nEUR|43|43\nPLN|10|10\n"
	if text != want {
		t.Fatalf("unexpected rates text\nwant: %q\n got: %q", want, text)
	}
}

func TestParseRatesTextRequiresAllTargetCurrencies(t *testing.T) {
	_, err := ParseRatesText([]monobank.CurrencyRate{
		{CurrencyCodeA: 840, CurrencyCodeB: 980, RateBuy: 40, RateSell: 41},
	})
	if err == nil {
		t.Fatal("expected error for missing currencies")
	}
}

func TestServiceWrapsProviderError(t *testing.T) {
	provider := &fakeProvider{err: errors.New("network down")}
	service := NewService(provider, 5*time.Minute)

	_, err := service.RatesText(context.Background())
	if !errors.Is(err, ErrUpstreamUnavailable) {
		t.Fatalf("expected ErrUpstreamUnavailable, got %v", err)
	}
}

func TestServiceWrapsInvalidRates(t *testing.T) {
	provider := &fakeProvider{
		rates: []monobank.CurrencyRate{
			{CurrencyCodeA: 840, CurrencyCodeB: 980, RateBuy: 40, RateSell: 41},
		},
	}
	service := NewService(provider, 5*time.Minute)

	_, err := service.RatesText(context.Background())
	if !errors.Is(err, ErrInvalidRates) {
		t.Fatalf("expected ErrInvalidRates, got %v", err)
	}
}

func TestServiceCachesRates(t *testing.T) {
	provider := &fakeProvider{
		rates: []monobank.CurrencyRate{
			{CurrencyCodeA: 840, CurrencyCodeB: 980, RateBuy: 40, RateSell: 41},
			{CurrencyCodeA: 978, CurrencyCodeB: 980, RateBuy: 43, RateSell: 44},
			{CurrencyCodeA: 985, CurrencyCodeB: 980, RateBuy: 10, RateSell: 11},
		},
	}
	service := NewService(provider, 5*time.Minute)
	currentTime := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return currentTime }

	first, err := service.RatesText(context.Background())
	if err != nil {
		t.Fatalf("first RatesText returned error: %v", err)
	}

	second, err := service.RatesText(context.Background())
	if err != nil {
		t.Fatalf("second RatesText returned error: %v", err)
	}

	if first != second {
		t.Fatalf("expected cached response to match first response")
	}
	if provider.calls != 1 {
		t.Fatalf("expected provider to be called once, got %d", provider.calls)
	}

	currentTime = currentTime.Add(6 * time.Minute)
	if _, err := service.RatesText(context.Background()); err != nil {
		t.Fatalf("third RatesText returned error: %v", err)
	}
	if provider.calls != 2 {
		t.Fatalf("expected provider to be called again after ttl, got %d", provider.calls)
	}
}
