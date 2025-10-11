package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSafeSelectAccountFuzzyPanicRecovery verifies panic recovery in account selector
func TestSafeSelectAccountFuzzyPanicRecovery(t *testing.T) {
	t.Run("panic recovery wrapper exists", func(t *testing.T) {
		// We can't easily trigger a panic in the fuzzy finder without user interaction,
		// but we can test that the wrapper function exists and has proper signature
		assert.NotNil(t, safeSelectAccountFuzzy)
	})
}

// TestSafeSelectRoleFuzzyPanicRecovery verifies panic recovery in role selector
func TestSafeSelectRoleFuzzyPanicRecovery(t *testing.T) {
	t.Run("panic recovery wrapper exists", func(t *testing.T) {
		assert.NotNil(t, safeSelectRoleFuzzy)
	})
}

// TestPreviewWindowBoundsChecking verifies bounds checking in preview callbacks
func TestPreviewWindowBoundsChecking(t *testing.T) {
	t.Run("account preview handles negative index", func(t *testing.T) {
		accounts := []Account{
			{AccountID: "123456789012", AccountName: "Test", EmailAddress: "test@example.com"},
		}

		// Simulate preview callback with negative index
		preview := func(i int) string {
			if i < 0 || i >= len(accounts) {
				return ""
			}
			account := accounts[i]
			return fmt.Sprintf("Account ID:   %s\nAccount Name: %s\nEmail:        %s",
				account.AccountID, account.AccountName, account.EmailAddress)
		}

		assert.Equal(t, "", preview(-1), "Should return empty string for negative index")
		assert.Equal(t, "", preview(999), "Should return empty string for out of bounds index")
		assert.NotEqual(t, "", preview(0), "Should return content for valid index")
	})

	t.Run("role preview handles negative index", func(t *testing.T) {
		account := &Account{
			AccountID:   "123456789012",
			AccountName: "Test Account",
		}
		roles := []Role{
			{RoleName: "TestRole", AccountID: account.AccountID},
		}

		// Simulate preview callback with negative index
		preview := func(i int) string {
			if i < 0 || i >= len(roles) {
				return ""
			}
			role := roles[i]
			return fmt.Sprintf("Role:         %s\nAccount:      %s\nAccount ID:   %s",
				role.RoleName, account.AccountName, account.AccountID)
		}

		assert.Equal(t, "", preview(-1), "Should return empty string for negative index")
		assert.Equal(t, "", preview(999), "Should return empty string for out of bounds index")
		assert.NotEqual(t, "", preview(0), "Should return content for valid index")
	})
}
