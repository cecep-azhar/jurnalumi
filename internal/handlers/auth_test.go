package handlers_test

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
	"time"

	"github.com/cecep-azhar/jurnalumi/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func TestResetPasswordExpiry(t *testing.T) {
	expired := time.Now().Add(-10 * time.Minute)
	user := models.User{
		Email:        "test@example.com",
		ResetToken:   "sample_token",
		ResetExpires: &expired,
	}

	if user.ResetExpires == nil || user.ResetExpires.Before(time.Now()) {
		// correctly expired
	} else {
		t.Fatalf("expected token to be expired")
	}

	valid := time.Now().Add(1 * time.Hour)
	user.ResetExpires = &valid
	if user.ResetExpires == nil || user.ResetExpires.Before(time.Now()) {
		t.Fatalf("expected token to be valid")
	}
}

func TestResetPasswordHashing(t *testing.T) {
	newPassword := "NewSecretPassword123"
	bytes := make([]byte, 32)
	rand.Read(bytes)
	token := hex.EncodeToString(bytes)

	if len(token) != 64 {
		t.Fatalf("expected token length 64, got %d", len(token))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(newPassword)); err != nil {
		t.Fatalf("password mismatch after bcrypt: %v", err)
	}
}
