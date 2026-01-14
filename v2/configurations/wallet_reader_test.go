package configurations

import (
	"bytes"
	"os"
	"testing"
)

// TestNewWalletFromReader verifies that NewWalletFromReader works the same as NewWallet
func TestNewWalletFromReader(t *testing.T) {
	// This test requires an actual wallet file to work
	// It's primarily a compilation test to ensure the API is correct

	// Create a mock reader with minimal valid wallet structure
	// In real usage, this would be a real cwallet.sso file
	mockData := []byte{161, 248, 78} // Magic bytes
	reader := bytes.NewReader(mockData)

	_, err := NewWalletFromReader(reader)
	// We expect an error because this is not a complete wallet
	// but it should not panic and should return a proper error
	if err == nil {
		t.Error("Expected error for incomplete wallet data, got nil")
	}
}

// TestReadFromBytesEquivalence verifies that read() and readFromBytes() produce same results
func TestReadFromBytesEquivalence(t *testing.T) {
	// Skip if no test wallet file is available
	testWalletPath := os.Getenv("TEST_WALLET_PATH")
	if testWalletPath == "" {
		t.Skip("Skipping test: TEST_WALLET_PATH environment variable not set")
	}

	// Test with file-based loading
	wallet1, err := NewWallet(testWalletPath)
	if err != nil {
		t.Fatalf("NewWallet failed: %v", err)
	}

	// Test with reader-based loading
	data, err := os.ReadFile(testWalletPath)
	if err != nil {
		t.Fatalf("Failed to read wallet file: %v", err)
	}

	reader := bytes.NewReader(data)
	wallet2, err := NewWalletFromReader(reader)
	if err != nil {
		t.Fatalf("NewWalletFromReader failed: %v", err)
	}

	// Compare results
	if len(wallet1.Certificates) != len(wallet2.Certificates) {
		t.Errorf("Certificate count mismatch: got %d, want %d",
			len(wallet2.Certificates), len(wallet1.Certificates))
	}

	if len(wallet1.credentials) != len(wallet2.credentials) {
		t.Errorf("Credentials count mismatch: got %d, want %d",
			len(wallet2.credentials), len(wallet1.credentials))
	}
}

// TestSetWalletFromReaderErrorHandling tests error cases
func TestNewWalletFromReaderErrorHandling(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "invalid magic bytes",
			data:    []byte{0, 0, 0},
			wantErr: true,
		},
		{
			name:    "incomplete header",
			data:    []byte{161, 248, 78, 54},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bytes.NewReader(tt.data)
			_, err := NewWalletFromReader(reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWalletFromReader() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
