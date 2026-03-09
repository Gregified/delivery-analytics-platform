package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func buildGatewayForTest(t *testing.T, key string, ordersHandler http.HandlerFunc, analyticsHandler http.HandlerFunc) http.Handler {
	t.Helper()

	ordersServer := httptest.NewServer(ordersHandler)
	t.Cleanup(ordersServer.Close)
	analyticsServer := httptest.NewServer(analyticsHandler)
	t.Cleanup(analyticsServer.Close)

	ordersURL, err := url.Parse(ordersServer.URL)
	if err != nil {
		t.Fatalf("parse orders URL: %v", err)
	}
	analyticsURL, err := url.Parse(analyticsServer.URL)
	if err != nil {
		t.Fatalf("parse analytics URL: %v", err)
	}

	return newGatewayHandler(key, ordersURL, analyticsURL)
}

func TestCORSPreflightReturns200AndHeaders(t *testing.T) {
	handler := buildGatewayForTest(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("orders upstream should not be called for OPTIONS")
	}, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("analytics upstream should not be called for OPTIONS")
	})

	req := httptest.NewRequest(http.MethodOptions, "/orders", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if got := res.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("expected CORS allow origin '*', got %q", got)
	}
	if got := res.Header.Get("Access-Control-Allow-Headers"); !strings.Contains(got, "X-API-Key") {
		t.Fatalf("expected allow headers to contain X-API-Key, got %q", got)
	}
}

func TestUnauthorizedRequestIsRejected(t *testing.T) {
	handler := buildGatewayForTest(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("orders upstream should not be called for unauthorized request")
	}, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("analytics upstream should not be called for unauthorized request")
	})

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.Header.Set("X-API-Key", "wrong-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
}

func TestOrdersRouteProxiesToOrdersService(t *testing.T) {
	handler := buildGatewayForTest(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/orders" {
			t.Fatalf("expected path /orders, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("orders-ok"))
	}, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("analytics upstream should not be called for /orders request")
	})

	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if string(body) != "orders-ok" {
		t.Fatalf("unexpected body: %q", string(body))
	}
}

func TestAnalyticsRouteProxiesToAnalyticsService(t *testing.T) {
	handler := buildGatewayForTest(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("orders upstream should not be called for /analytics request")
	}, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/analytics/summary" {
			t.Fatalf("expected path /analytics/summary, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("analytics-ok"))
	})

	req := httptest.NewRequest(http.MethodGet, "/analytics/summary", nil)
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	if string(body) != "analytics-ok" {
		t.Fatalf("unexpected body: %q", string(body))
	}
}

func TestUnknownRouteReturns404WithoutCallingUpstreams(t *testing.T) {
	handler := buildGatewayForTest(t, "test-key", func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("orders upstream should not be called for unknown route")
	}, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("analytics upstream should not be called for unknown route")
	})

	req := httptest.NewRequest(http.MethodGet, "/not-a-route", nil)
	req.Header.Set("X-API-Key", "test-key")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	res := rec.Result()
	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
	if !strings.Contains(string(body), "Route not found") {
		t.Fatalf("expected body to contain 'Route not found', got %q", string(body))
	}
}
