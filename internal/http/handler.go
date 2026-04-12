package http

import (
	"config-analyzer/internal/model"
	"config-analyzer/internal/service"
	"encoding/json"
	"net/http"
	"strings"

	"go.yaml.in/yaml/v3"
)

type Status string

const (
	StatusClean   Status = "clean"
	StatusIssued  Status = "issues"
	StatusError   Status = "error"
)

type AnalyzeResponse struct {
	Status       Status        `json:"status"`
	Issues       []model.Issue `json:"issues,omitempty"`
	ErrorMessage string        `json:"error_message,omitempty"`
}

type Handler struct {
	analyzer service.Analyzer
}

func NewHandler(a service.Analyzer) *Handler {
	return &Handler{analyzer: a}
}

func (h *Handler) Analyze(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cfg := new(model.Config)
	contentType := r.Header.Get("Content-Type")

	var err error
	switch {
	case strings.Contains(contentType, "yaml"), strings.Contains(contentType, "yml"):
		err = yaml.NewDecoder(r.Body).Decode(cfg)
	default:
		err = json.NewDecoder(r.Body).Decode(cfg)
	}

	if err != nil {
		resp := AnalyzeResponse{
			Status:       StatusError,
			ErrorMessage: err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	issues := h.analyzer.Analyze(cfg)

	status := StatusClean
	if len(issues) > 0 {
		status = StatusIssued
	}

	resp := AnalyzeResponse{
		Status: status,
		Issues: issues,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
