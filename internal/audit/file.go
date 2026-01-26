package audit

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/dmnAlex/shortener/internal/model"
	"github.com/pkg/errors"
)

type FileAuditor struct {
	file    *os.File
	encoder *json.Encoder
	mu      sync.Mutex
}

func NewFileAuditor(path string) (*FileAuditor, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "open file")
	}

	return &FileAuditor{
		file:    file,
		encoder: json.NewEncoder(file),
	}, nil
}

func (a *FileAuditor) Audit(event model.AuditEvent) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return errors.Wrap(a.encoder.Encode(event), "encode event")
}

func (a *FileAuditor) Close() error {
	if a.file != nil {
		return a.file.Close()
	}

	return nil
}
