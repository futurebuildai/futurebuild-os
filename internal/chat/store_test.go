package chat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTx implements pgx.Tx for testing transactional behavior.
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockTx) Commit(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Rollback(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockTx) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

// Implement remaining pgx.Tx interface methods (no-op for tests)
func (m *MockTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (m *MockTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	return nil
}

func (m *MockTx) LargeObjects() pgx.LargeObjects {
	return pgx.LargeObjects{}
}

func (m *MockTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (m *MockTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return nil, nil
}

func (m *MockTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return nil
}

func (m *MockTx) Conn() *pgx.Conn {
	return nil
}

// MockTransactor implements Transactor interface for testing.
type MockTransactor struct {
	mock.Mock
	MockTxInstance *MockTx
}

func (m *MockTransactor) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockTransactor) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

func TestPgxMessageStore_SaveMessage_Success(t *testing.T) {
	// Arrange
	mockTx := &MockTx{}
	mockTransactor := &MockTransactor{}
	store := NewPgxMessageStore(mockTransactor)

	userID := uuid.New()
	msg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: uuid.New(),
		UserID:    &userID,
		Role:      types.ChatRoleUser,
		Content:   "Test Message",
		CreatedAt: time.Now(),
		// No ToolCalls
	}

	// Expect Begin to return our mock transaction
	mockTransactor.On("Begin", mock.Anything).Return(mockTx, nil)

	// Expect Exec on transaction for chat_messages insert
	mockTx.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(pgconn.CommandTag{}, nil).
		Run(func(args mock.Arguments) {
			sql := args.String(1)
			assert.Contains(t, sql, "INSERT INTO chat_messages")
		})

	// Expect Commit
	mockTx.On("Commit", mock.Anything).Return(nil)

	// Expect Rollback (deferred, no-op after commit)
	mockTx.On("Rollback", mock.Anything).Return(nil)

	// Act
	err := store.SaveMessage(context.Background(), msg)

	// Assert
	require.NoError(t, err)
	mockTransactor.AssertExpectations(t)
	mockTx.AssertExpectations(t)
}

func TestPgxMessageStore_SaveMessage_WithToolCalls(t *testing.T) {
	// Arrange
	mockTx := &MockTx{}
	mockTransactor := &MockTransactor{}
	store := NewPgxMessageStore(mockTransactor)

	toolUserID := uuid.New()
	msg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: uuid.New(),
		UserID:    &toolUserID,
		Role:      types.ChatRoleModel,
		Content:   "Let me check the schedule",
		ToolCalls: []types.ToolCall{
			{
				ID:       "call_123",
				Name:     "get_schedule",
				Args:     map[string]interface{}{"project_id": "abc"},
				Response: "Schedule retrieved",
			},
			{
				ID:       "call_456",
				Name:     "update_task",
				Args:     map[string]interface{}{"task_id": "xyz"},
				Response: "Task updated",
			},
		},
		CreatedAt: time.Now(),
	}

	// Expect Begin
	mockTransactor.On("Begin", mock.Anything).Return(mockTx, nil)

	// Expect 3 Exec calls: 1 for message + 2 for tool calls
	mockTx.On("Exec", mock.Anything, mock.AnythingOfType("string"), mock.Anything).
		Return(pgconn.CommandTag{}, nil).Times(3)

	// Expect Commit
	mockTx.On("Commit", mock.Anything).Return(nil)

	// Expect Rollback (deferred)
	mockTx.On("Rollback", mock.Anything).Return(nil)

	// Act
	err := store.SaveMessage(context.Background(), msg)

	// Assert
	require.NoError(t, err)
	mockTransactor.AssertExpectations(t)
	mockTx.AssertNumberOfCalls(t, "Exec", 3)
}

func TestPgxMessageStore_SaveMessage_RollbackOnToolInsertFailure(t *testing.T) {
	// Arrange
	mockTx := &MockTx{}
	mockTransactor := &MockTransactor{}
	store := NewPgxMessageStore(mockTransactor)

	rollbackUserID := uuid.New()
	msg := models.ChatMessage{
		ID:        uuid.New(),
		ProjectID: uuid.New(),
		UserID:    &rollbackUserID,
		Role:      types.ChatRoleModel,
		Content:   "Let me process this",
		ToolCalls: []types.ToolCall{
			{
				ID:   "call_fail",
				Name: "failing_tool",
				Args: map[string]interface{}{},
			},
		},
		CreatedAt: time.Now(),
	}

	// Expect Begin
	mockTransactor.On("Begin", mock.Anything).Return(mockTx, nil)

	// First Exec succeeds (chat_messages insert)
	mockTx.On("Exec", mock.Anything, mock.MatchedBy(func(sql string) bool {
		return true // Match any SQL for first call
	}), mock.Anything).
		Return(pgconn.CommandTag{}, nil).Once()

	// Second Exec fails (tool_usage insert)
	mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).
		Return(pgconn.CommandTag{}, errors.New("db error on tool insert")).Once()

	// Rollback should be called due to error (deferred)
	mockTx.On("Rollback", mock.Anything).Return(nil)

	// Act
	err := store.SaveMessage(context.Background(), msg)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "chat_tool_usage failed")
	assert.Contains(t, err.Error(), "failing_tool")
	mockTransactor.AssertExpectations(t)
}

func TestPgxMessageStore_SaveMessage_BeginError(t *testing.T) {
	// Arrange
	mockTransactor := &MockTransactor{}
	store := NewPgxMessageStore(mockTransactor)

	// Expect Begin to fail
	mockTransactor.On("Begin", mock.Anything).Return(nil, errors.New("connection error"))

	// Act
	err := store.SaveMessage(context.Background(), models.ChatMessage{})

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to begin transaction")
	mockTransactor.AssertExpectations(t)
}

func TestPgxMessageStore_SaveMessage_MessageInsertError(t *testing.T) {
	// Arrange
	mockTx := &MockTx{}
	mockTransactor := &MockTransactor{}
	store := NewPgxMessageStore(mockTransactor)

	// Expect Begin
	mockTransactor.On("Begin", mock.Anything).Return(mockTx, nil)

	// Expect Exec to fail on message insert
	mockTx.On("Exec", mock.Anything, mock.Anything, mock.Anything).
		Return(pgconn.CommandTag{}, errors.New("db error"))

	// Rollback due to error
	mockTx.On("Rollback", mock.Anything).Return(nil)

	// Act
	err := store.SaveMessage(context.Background(), models.ChatMessage{})

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db insert chat_messages failed")
	mockTransactor.AssertExpectations(t)
}
