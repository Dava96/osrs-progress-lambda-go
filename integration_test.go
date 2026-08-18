//go:build integration

package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/joho/godotenv"
)
import _ "github.com/joho/godotenv/autoload"

func TestLocalLambdaFlow(t *testing.T) {
	if err := godotenv.Load(".env.integration"); err != nil {
		t.Fatalf("failed to load the integration env vars: %v", err)
	}

	if err := handleRequest(context.Background(), json.RawMessage(`{}`)); err != nil {
		t.Fatalf("local Lambda flow failed: %v", err)
	}
}
