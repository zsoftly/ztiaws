package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/aws/aws-sdk-go-v2/service/sso/types"
	"github.com/stretchr/testify/assert"
)

// mockSSOClient is a mock implementation of the sso.ListAccountsAPIClient and sso.ListAccountRolesAPIClient interfaces.
type mockSSOClient struct {
	ListAccountsPages         [][]types.AccountInfo
	ListAccountRolesPages     [][]types.RoleInfo
	ListAccountsErr           error
	ListAccountRolesErr       error
	listAccountsCallCount     int
	listAccountRolesCallCount int
}

func (m *mockSSOClient) ListAccounts(ctx context.Context, params *sso.ListAccountsInput, optFns ...func(*sso.Options)) (*sso.ListAccountsOutput, error) {
	if m.ListAccountsErr != nil {
		return nil, m.ListAccountsErr
	}
	if m.listAccountsCallCount >= len(m.ListAccountsPages) {
		// Return empty list and no next token when all pages are exhausted
		return &sso.ListAccountsOutput{AccountList: []types.AccountInfo{}, NextToken: nil}, nil
	}

	page := m.ListAccountsPages[m.listAccountsCallCount]
	m.listAccountsCallCount++

	var nextToken *string
	if m.listAccountsCallCount < len(m.ListAccountsPages) {
		nextToken = aws.String("next-page")
	}

	return &sso.ListAccountsOutput{
		AccountList: page,
		NextToken:   nextToken,
	}, nil
}

func (m *mockSSOClient) ListAccountRoles(ctx context.Context, params *sso.ListAccountRolesInput, optFns ...func(*sso.Options)) (*sso.ListAccountRolesOutput, error) {
	if m.ListAccountRolesErr != nil {
		return nil, m.ListAccountRolesErr
	}
	if m.listAccountRolesCallCount >= len(m.ListAccountRolesPages) {
		// Return empty list and no next token when all pages are exhausted
		return &sso.ListAccountRolesOutput{RoleList: []types.RoleInfo{}, NextToken: nil}, nil
	}

	page := m.ListAccountRolesPages[m.listAccountRolesCallCount]
	m.listAccountRolesCallCount++

	var nextToken *string
	if m.listAccountRolesCallCount < len(m.ListAccountRolesPages) {
		nextToken = aws.String("next-page")
	}

	return &sso.ListAccountRolesOutput{
		RoleList:  page,
		NextToken: nextToken,
	}, nil
}

func TestListAccountsWithPagination(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	tests := []struct {
		name          string
		pages         [][]types.AccountInfo
		expectedCount int
		expectErr     bool
	}{
		{
			name:          "single page of accounts",
			pages:         [][]types.AccountInfo{{{AccountId: aws.String("1"), AccountName: aws.String("Acc1")}, {AccountId: aws.String("2"), AccountName: aws.String("Acc2")}}},
			expectedCount: 2,
			expectErr:     false,
		},
		{
			name: "multiple pages of accounts",
			pages: [][]types.AccountInfo{
				{{AccountId: aws.String("1"), AccountName: aws.String("Acc1")}},
				{{AccountId: aws.String("2"), AccountName: aws.String("Acc2")}},
				{{AccountId: aws.String("3"), AccountName: aws.String("Acc3")}},
			},
			expectedCount: 3,
			expectErr:     false,
		},
		{
			name:          "no accounts",
			pages:         [][]types.AccountInfo{},
			expectedCount: 0,
			expectErr:     false,
		},
		{
			name:          "empty pages",
			pages:         [][]types.AccountInfo{{}, {}},
			expectedCount: 0,
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSSOClient{
				ListAccountsPages: tt.pages,
			}

			accounts, err := manager.listAccounts(ctx, mockClient, "fake-token")

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, accounts, tt.expectedCount)
			}
		})
	}
}

func TestListAccountRolesWithPagination(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()

	tests := []struct {
		name          string
		pages         [][]types.RoleInfo
		expectedCount int
		expectErr     bool
	}{
		{
			name:          "single page of roles",
			pages:         [][]types.RoleInfo{{{RoleName: aws.String("Role1")}, {RoleName: aws.String("Role2")}}},
			expectedCount: 2,
			expectErr:     false,
		},
		{
			name: "multiple pages of roles",
			pages: [][]types.RoleInfo{
				{{RoleName: aws.String("Role1")}},
				{{RoleName: aws.String("Role2")}},
				{{RoleName: aws.String("Role3")}},
			},
			expectedCount: 3,
			expectErr:     false,
		},
		{
			name:          "no roles",
			pages:         [][]types.RoleInfo{},
			expectedCount: 0,
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &mockSSOClient{
				ListAccountRolesPages: tt.pages,
			}

			roles, err := manager.listAccountRoles(ctx, mockClient, "fake-token", "fake-account-id")

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, roles, tt.expectedCount)
			}
		})
	}
}

func TestListAccountsError(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()
	mockClient := &mockSSOClient{
		ListAccountsErr: fmt.Errorf("AWS error"),
	}

	_, err := manager.listAccounts(ctx, mockClient, "fake-token")
	assert.Error(t, err)
}

func TestListAccountRolesError(t *testing.T) {
	manager := NewManager()
	ctx := context.Background()
	mockClient := &mockSSOClient{
		ListAccountRolesErr: fmt.Errorf("AWS error"),
	}

	_, err := manager.listAccountRoles(ctx, mockClient, "fake-token", "fake-account-id")
	assert.Error(t, err)
}
