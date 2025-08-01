package utils

import (
	"strings"
	"testing"
)

func TestParseCurrencyPair_Valid(t *testing.T) {
	base, quote, err := ParseCurrencyPair("usd/eur")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != "USD" || quote != "EUR" {
		t.Errorf("expected USD/EUR, got %s/%s", base, quote)
	}
}

func TestParseCurrencyPair_ValidUpperCase(t *testing.T) {
	base, quote, err := ParseCurrencyPair("GBP/JPY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != "GBP" || quote != "JPY" {
		t.Errorf("expected GBP/JPY, got %s/%s", base, quote)
	}
}

func TestParseCurrencyPair_InvalidLength(t *testing.T) {
	_, _, err := ParseCurrencyPair("usd/eu")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid currency pair format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseCurrencyPair_MissingSlash(t *testing.T) {
	_, _, err := ParseCurrencyPair("usdEUR")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid currency pair format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseCurrencyPair_ExtraSlash(t *testing.T) {
	_, _, err := ParseCurrencyPair("usd/e/r")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid currency pair format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseCurrencyPair_WrongSegmentLength(t *testing.T) {
	_, _, err := ParseCurrencyPair("us/euro")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid currency pair format") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestParseCurrencyPair_Empty(t *testing.T) {
	_, _, err := ParseCurrencyPair("")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "invalid currency pair format") {
		t.Errorf("unexpected error message: %v", err)
	}
}
