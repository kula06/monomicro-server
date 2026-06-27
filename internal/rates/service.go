package rates

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"monomicro-server/internal/monobank"
)

const uahCurrencyCode = 980

var (
	ErrUpstreamUnavailable = errors.New("upstream unavailable")
	ErrInvalidRates        = errors.New("invalid rates data")
)

var targetCurrencies = []targetCurrency{
	{Code: 840, Symbol: "USD"},
	{Code: 978, Symbol: "EUR"},
	{Code: 985, Symbol: "PLN"},
}

type targetCurrency struct {
	Code   int
	Symbol string
}

type Provider interface {
	FetchCurrencyRates(ctx context.Context) ([]monobank.CurrencyRate, error)
}

type Service struct {
	provider Provider
	ttl      time.Duration
	now      func() time.Time

	mu        sync.Mutex
	cached    string
	expiresAt time.Time
}

func NewService(provider Provider, ttl time.Duration) *Service {
	return &Service{
		provider: provider,
		ttl:      ttl,
		now:      time.Now,
	}
}

func (s *Service) RatesText(ctx context.Context) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if s.cached != "" && now.Before(s.expiresAt) {
		return s.cached, nil
	}

	sourceRates, err := s.provider.FetchCurrencyRates(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrUpstreamUnavailable, err)
	}

	text, err := ParseRatesText(sourceRates)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidRates, err)
	}

	s.cached = text
	s.expiresAt = now.Add(s.ttl)

	return text, nil
}

func ParseRatesText(sourceRates []monobank.CurrencyRate) (string, error) {
	lines := make([]string, 0, len(targetCurrencies))
	byCode := make(map[int]monobank.CurrencyRate, len(sourceRates))

	for _, rate := range sourceRates {
		if rate.CurrencyCodeB != uahCurrencyCode {
			continue
		}

		if rate.RateBuy <= 0 && rate.RateSell <= 0 && rate.RateCross <= 0 {
			continue
		}

		byCode[rate.CurrencyCodeA] = rate
	}

	var missing []string
	for _, currency := range targetCurrencies {
		rate, ok := byCode[currency.Code]
		if !ok {
			missing = append(missing, currency.Symbol)
			continue
		}

		buy, sell := normalizeBuySell(rate)
		if buy <= 0 || sell <= 0 {
			missing = append(missing, currency.Symbol)
			continue
		}

		lines = append(lines, fmt.Sprintf("%s|%s|%s", currency.Symbol, formatRate(buy), formatRate(sell)))
	}

	if len(missing) > 0 {
		return "", fmt.Errorf("missing rates for: %s", strings.Join(missing, ", "))
	}

	if len(lines) == 0 {
		return "", errors.New("no rates available")
	}

	return strings.Join(lines, "\n") + "\n", nil
}

func normalizeBuySell(rate monobank.CurrencyRate) (float64, float64) {
	buy := rate.RateBuy
	sell := rate.RateSell

	if buy <= 0 && rate.RateCross > 0 {
		buy = rate.RateCross
	}

	if sell <= 0 && rate.RateCross > 0 {
		sell = rate.RateCross
	}

	return buy, sell
}

func formatRate(rate float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.6f", rate), "0"), ".")
}
