package main

import (
	"os"
	"testing"

	jwch "github.com/west2-online/jwch"
)

func TestIntegrationLoginAndQuery(t *testing.T) {
	id := os.Getenv("FZU_JWCH_ID")
	password := os.Getenv("FZU_JWCH_PASSWORD")
	if id == "" || password == "" {
		t.Skip("set FZU_JWCH_ID and FZU_JWCH_PASSWORD to run integration test")
	}

	student := jwch.NewStudent().WithUser(id, password)
	if err := student.Login(); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if err := student.CheckSession(); err != nil {
		t.Fatalf("CheckSession() error = %v", err)
	}
	if _, err := student.GetInfo(); err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}
}
