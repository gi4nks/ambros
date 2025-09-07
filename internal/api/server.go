package api

import (
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"github.com/gi4nks/ambros/internal/models"
)

// RepositoryInterface defines the repository methods needed by the API
type RepositoryInterface interface {
	Get(id string) (*models.Command, error)
	GetAllCommands() ([]models.Command, error)
}

// Server represents the API server
type Server struct {
	logger     *zap.Logger
	repository RepositoryInterface
}

// NewServer creates a new API server
func NewServer(repo RepositoryInterface, logger *zap.Logger) *Server {
	return &Server{
		logger:     logger,
		repository: repo,
	}
}

// SetupRoutes sets up the HTTP routes
func (s *Server) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/commands", s.handleCommands)
	mux.HandleFunc("/commands/", s.handleCommand)
	mux.HandleFunc("/health", s.handleHealth)

	return mux
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		commands, err := s.repository.GetAllCommands()
		if err != nil {
			s.logger.Error("Failed to get commands", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(commands); err != nil {
			s.logger.Error("Failed to encode commands", zap.Error(err))
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleCommand(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/commands/"):]
	if id == "" {
		http.Error(w, "Command ID required", http.StatusBadRequest)
		return
	}

	command, err := s.repository.Get(id)
	if err != nil {
		s.logger.Error("Failed to get command",
			zap.String("id", id),
			zap.Error(err))
		http.Error(w, "Command not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(command); err != nil {
		s.logger.Error("Failed to encode command", zap.Error(err))
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{"status": "healthy"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode health response", zap.Error(err))
	}
}
