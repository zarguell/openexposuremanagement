package database

import (
	"testing"
)

func TestConnectWithInvalidURL(t *testing.T) {
	_, err := Connect("postgres://invalid:connection@string")
	if err == nil {
		t.Error("Expected error when connecting with invalid URL")
	}
}
