package audit

import (
	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/pkg/errors"
)

type Auditor interface {
	Audit(event model.AuditEvent) error
	Close() error
}

type AuditManager struct {
	auditors []Auditor
}

func NewAuditManager() *AuditManager {
	return &AuditManager{
		auditors: make([]Auditor, 0),
	}
}

func (m *AuditManager) Subscribe(auditor Auditor) {
	m.auditors = append(m.auditors, auditor)
}

func (m *AuditManager) Notify(event model.AuditEvent) {
	for _, a := range m.auditors {
		if err := a.Audit(event); err != nil {
			err = errors.Wrap(err, "audit event")
			logger.Log.Error(err.Error())
		}
	}
}

func (m *AuditManager) Close() {
	for _, a := range m.auditors {
		if err := a.Close(); err != nil {
			err = errors.Wrap(err, "close auditor")
			logger.Log.Error(err.Error())
		}
	}
}
