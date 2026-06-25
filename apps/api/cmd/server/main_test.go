package main

import (
	"errors"
	"strings"
	"testing"

	"continuum/api/internal/repository"
)

func TestConnectToDatabaseWithRetryReturnsErrorAfterFailures(t *testing.T) {
	originalFactory := databaseConnectFactory
	originalAttempts := databaseRetryAttempts
	originalDelay := databaseRetryDelay

	databaseConnectFactory = func(string) (*repository.Database, error) {
		return nil, errors.New("boom")
	}
	databaseRetryAttempts = 1
	databaseRetryDelay = 0

	defer func() {
		databaseConnectFactory = originalFactory
		databaseRetryAttempts = originalAttempts
		databaseRetryDelay = originalDelay
	}()

	_, err := connectToDatabaseWithRetry("postgres://example.invalid/db")
	if err == nil {
		t.Fatal("expected an error from the retry loop")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected wrapped connection error, got %v", err)
	}
}

func TestApplyProductionMigrationsSkipsNilDatabase(t *testing.T) {
	if err := applyProductionMigrations(nil); err != nil {
		t.Fatalf("expected nil database to be ignored, got %v", err)
	}
}
