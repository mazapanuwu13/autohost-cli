package agentconfig

import (
	"testing"
)

func TestCleanGRPCAddress(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://grpc.autohst.dev", "grpc.autohst.dev:443"},
		{"https://grpc.autohst.dev/", "grpc.autohst.dev:443"},
		{"grpc.autohst.dev:443", "grpc.autohst.dev:443"},
		{"grpc.autohst.dev", "grpc.autohst.dev:443"},
		{"http://localhost:9090", "localhost:9090"},
		{":9090", ":9090"},
		{"192.168.1.50:9090", "192.168.1.50:9090"},
		{"", ""},
	}

	for _, tt := range tests {
		got := CleanGRPCAddress(tt.input)
		if got != tt.expected {
			t.Errorf("CleanGRPCAddress(%q) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
