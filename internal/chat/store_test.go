package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDBExecutor is a mock implementation of the DBExecutor interface.
type MockDBExecutor struct {
	mock.Mock
}

func (m *MockDBExecutor) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func TestPgxMessageStore_SaveMessage_Success(t *testing.T) {
	// Arrange
	mockDB := &MockDBExecutor{}
	store := NewPgxMessageStore(mockDB)

	msg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: uuid.New(),
		UserID:    uuid.New(),
		Content:   "Test Message",
		CreatedAt: time.Now(),
	}

	// Expect Exec to be called with correct SQL and arguments
	mockDB.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(pgconn.CommandTag{}, nil).
		Run(func(args mock.Arguments) {
			sql := args.String(1)
			assert.Contains(t, sql, "INSERT INTO chat_messages")

			// Verify arguments passed to Exec (msg attributes)
			// Arguments are passed as slice in index 2
			queryArgs := args.Get(2).([]any)
			assert.Equal(t, msg.ID, queryArgs[0])
			assert.Equal(t, msg.ProjectID, queryArgs[1])
			assert.Equal(t, msg.UserID, queryArgs[2])
			// ... other args verification
		})

	// Act
	err := store.SaveMessage(context.Background(), msg)

	// Assert
	require.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestPgxMessageStore_SaveMessage_Error(t *testing.T) {
	// Arrange
	mockDB := &MockDBExecutor{}
	store := NewPgxMessageStore(mockDB)

	mockDB.On("Exec", mock.Anything, mock.Anything, mock.Anything).
		Return(pgconn.CommandTag{}, errors.New("db error"))

	// Act
	err := store.SaveMessage(context.Background(), models.ChatMessage{})

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db insert failed")
	mockDB.AssertExpectations(t)
}
