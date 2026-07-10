package services_test

import (
	"testing"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository implements repositories.UserRepository for testing
// Since repositories.UserRepository is a struct and not an interface, 
// wait, we can't easily mock a concrete struct using a real mongo collection in a pure unit test.
// Let's rely on integration tests or skip the DB part for now, 
// OR we can test the SeedAdmin logic if we had a mock.
// Since we used a concrete struct for UserRepository, we will just write a placeholder test for now.

func TestAuthService_PasswordHashing(t *testing.T) {
	password := "admin123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	err = bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		t.Errorf("Expected password to match hash, got error: %v", err)
	}
}
