package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "secret123"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	
	if hash == "" {
		t.Error("Hash should not be empty")
	}
	
	if hash == password {
		t.Error("Hash should not equal plain password")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "secret123"
	hash, _ := HashPassword(password)
	
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword should return true for correct password")
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	password := "secret123"
	hash, _ := HashPassword(password)
	
	if VerifyPassword("wrongpassword", hash) {
		t.Error("VerifyPassword should return false for incorrect password")
	}
}

func TestHashMySQLNativePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{
			name:     "empty password",
			password: "",
			want:     "",
		},
		{
			name:     "non-empty password",
			password: "password",
			want:     "*2470C0C06DEE42FD1618BB99005ADCA2EC9D1E19",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashMySQLNativePassword(tt.password)
			if got != tt.want {
				t.Errorf("HashMySQLNativePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordHashUniqueness(t *testing.T) {
	password := "secret123"
	
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)
	
	// bcrypt includes a salt, so hashes should be different
	if hash1 == hash2 {
		t.Error("Same password should produce different hashes due to salt")
	}
	
	// But both should verify correctly
	if !VerifyPassword(password, hash1) || !VerifyPassword(password, hash2) {
		t.Error("Both hashes should verify correctly")
	}
}
