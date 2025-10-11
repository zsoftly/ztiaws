package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateAccountSelectorWidth(t *testing.T) {
	tests := []struct {
		name          string
		accounts      []Account
		expectedMin   int
		expectedMax   int
		expectExactly int
		checkBounds   bool
	}{
		{
			name:          "empty accounts - should return minimum width",
			accounts:      []Account{},
			expectExactly: 80,
		},
		{
			name: "short account names - should return minimum width",
			accounts: []Account{
				{AccountID: "123456789012", AccountName: "Dev"},
				{AccountID: "123456789013", AccountName: "Test"},
			},
			expectExactly: 80,
		},
		{
			name: "medium length account names",
			accounts: []Account{
				{AccountID: "123456789012", AccountName: "Development Environment"},
				{AccountID: "123456789013", AccountName: "Production Environment"},
			},
			expectedMin: 80,
			expectedMax: 160,
			checkBounds: true,
		},
		{
			name: "very long account names - should be capped by terminal width",
			accounts: []Account{
				{
					AccountID:   "123456789012",
					AccountName: "Super Long Account Name That Should Definitely Exceed Any Reasonable Terminal Width For Testing Purposes And Then Some More To Be Sure",
				},
			},
			expectedMin: 80,
			checkBounds: true,
		},
		{
			name: "mixed length account names",
			accounts: []Account{
				{AccountID: "111111111111", AccountName: "Short"},
				{AccountID: "222222222222", AccountName: "Medium Length Name"},
				{AccountID: "333333333333", AccountName: "Very Long Account Name For Testing"},
			},
			expectedMin: 80,
			expectedMax: getTerminalWidth(),
			checkBounds: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width := calculateAccountSelectorWidth(tt.accounts)

			if tt.expectExactly > 0 {
				assert.Equal(t, tt.expectExactly, width, "Width should match expected exact value")
			}

			if tt.checkBounds {
				assert.GreaterOrEqual(t, width, tt.expectedMin, "Width should be >= minimum")
				if tt.expectedMax > 0 {
					assert.LessOrEqual(t, width, tt.expectedMax, "Width should be <= maximum")
				}
			}

			// Always check absolute bounds
			assert.GreaterOrEqual(t, width, 80, "Width should never be less than minimum (80)")
			assert.LessOrEqual(t, width, getTerminalWidth(), "Width should never exceed terminal width")
		})
	}
}

func TestCalculateRoleSelectorWidth(t *testing.T) {
	testAccount := &Account{
		AccountID:   "123456789012",
		AccountName: "Test Account",
	}

	tests := []struct {
		name          string
		roles         []Role
		account       *Account
		expectedMin   int
		expectedMax   int
		expectExactly int
		checkBounds   bool
	}{
		{
			name:        "empty roles - header determines width",
			roles:       []Role{},
			account:     testAccount,
			expectedMin: 80,
			expectedMax: getTerminalWidth(),
			checkBounds: true,
		},
		{
			name: "short role names - header may determine width",
			roles: []Role{
				{RoleName: "Admin", AccountID: "123456789012"},
				{RoleName: "User", AccountID: "123456789012"},
			},
			account:     testAccount,
			expectedMin: 80,
			expectedMax: getTerminalWidth(),
			checkBounds: true,
		},
		{
			name: "medium length role names",
			roles: []Role{
				{RoleName: "PowerUserAccess", AccountID: "123456789012"},
				{RoleName: "ReadOnlyAccess", AccountID: "123456789012"},
			},
			account:     testAccount,
			expectedMin: 80,
			expectedMax: getTerminalWidth(),
			checkBounds: true,
		},
		{
			name: "very long role names - should be capped by terminal width",
			roles: []Role{
				{
					RoleName:  "SuperLongRoleNameThatShouldDefinitelyExceedAnyReasonableTerminalWidthForTestingPurposesAndThenSomeMoreCharactersJustToBeSafe",
					AccountID: "123456789012",
				},
			},
			account:     testAccount,
			expectedMin: 80,
			checkBounds: true,
		},
		{
			name: "long account name affects width",
			roles: []Role{
				{RoleName: "Admin", AccountID: "123456789012"},
			},
			account: &Account{
				AccountID:   "123456789012",
				AccountName: "Very Long Account Name That Affects The Header Width Calculation And Should Be Handled Gracefully By The Dynamic Width Calculation",
			},
			expectedMin: 80,
			expectedMax: getTerminalWidth(),
			checkBounds: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width := calculateRoleSelectorWidth(tt.roles, tt.account)

			if tt.expectExactly > 0 {
				assert.Equal(t, tt.expectExactly, width, "Width should match expected exact value")
			}

			if tt.checkBounds {
				assert.GreaterOrEqual(t, width, tt.expectedMin, "Width should be >= minimum")
				if tt.expectedMax > 0 {
					assert.LessOrEqual(t, width, tt.expectedMax, "Width should be <= maximum")
				}
			}

			// Always check absolute bounds
			assert.GreaterOrEqual(t, width, 80, "Width should never be less than minimum (80)")
			assert.LessOrEqual(t, width, getTerminalWidth(), "Width should never exceed terminal width")
		})
	}
}

func TestWidthCalculationConsistency(t *testing.T) {
	t.Run("same input produces same output", func(t *testing.T) {
		accounts := []Account{
			{AccountID: "123456789012", AccountName: "Test Account One"},
			{AccountID: "123456789013", AccountName: "Test Account Two"},
		}

		width1 := calculateAccountSelectorWidth(accounts)
		width2 := calculateAccountSelectorWidth(accounts)

		assert.Equal(t, width1, width2, "Same input should produce consistent output")
	})

	t.Run("role width calculation is consistent", func(t *testing.T) {
		account := &Account{
			AccountID:   "123456789012",
			AccountName: "Test Account",
		}
		roles := []Role{
			{RoleName: "AdminRole", AccountID: "123456789012"},
			{RoleName: "UserRole", AccountID: "123456789012"},
		}

		width1 := calculateRoleSelectorWidth(roles, account)
		width2 := calculateRoleSelectorWidth(roles, account)

		assert.Equal(t, width1, width2, "Same input should produce consistent output")
	})
}
