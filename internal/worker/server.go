package worker

import (
	"context"
	"log/slog"

	"github.com/hibiken/asynq"
)

// Server wraps the asynq.Server
type Server struct {
	srv *asynq.Server
	mux *asynq.ServeMux
}

// NewServer creates a new worker server instance
func NewServer(redisAddr string, concurrency int, queues map[string]int) *Server {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: concurrency,
			Queues:      queues,
			ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
				slog.Error("task processing failed", "type", task.Type(), "payload", string(task.Payload()), "error", err)
			}),
		},
	)

	return &Server{
		srv: srv,
		mux: asynq.NewServeMux(),
	}
}

// RegisterHandler registers a handler for a specific task type
func (s *Server) RegisterHandler(pattern string, handler asynq.Handler) {
	s.mux.Handle(pattern, handler)
}

// RegisterHandlerFunc registers a handler function for a specific task type
func (s *Server) RegisterHandlerFunc(pattern string, handler func(context.Context, *asynq.Task) error) {
	s.mux.HandleFunc(pattern, handler)
}

// Start starts the worker server
func (s *Server) Run() error {
	slog.Info("Starting Asynq Worker Server...")
	return s.srv.Run(s.mux)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.srv.Shutdown()
}
