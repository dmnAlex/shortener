package audit

import (
	"sync"

	"github.com/dmnAlex/shortener/internal/logger"
	"github.com/dmnAlex/shortener/internal/model"
	"github.com/pkg/errors"
)

const (
	managerChanSize = 100
	wrapperChanSize = 10
)

type Auditor interface {
	Audit(event model.AuditEvent) error
	Close() error
}

type auditorWrapper struct {
	auditor Auditor
	ch      chan model.AuditEvent
}

type AuditManager struct {
	wrappers []*auditorWrapper
	notifyCh chan model.AuditEvent
	wg       sync.WaitGroup
	done     chan struct{}
}

func NewAuditManager() *AuditManager {
	m := &AuditManager{
		notifyCh: make(chan model.AuditEvent, managerChanSize),
		done:     make(chan struct{}),
	}

	m.wg.Add(1)
	go m.worker()

	return m
}

func (m *AuditManager) worker() {
	defer m.wg.Done()
	for {
		select {
		case event := <-m.notifyCh:
			for _, w := range m.wrappers {
				w.ch <- event
			}
		case <-m.done:
			return
		}
	}
}

func (m *AuditManager) Subscribe(auditor Auditor) {
	ch := make(chan model.AuditEvent, wrapperChanSize)
	w := &auditorWrapper{
		auditor: auditor,
		ch:      ch,
	}
	m.wrappers = append(m.wrappers, w)
	m.wg.Add(1)

	go func() {
		defer m.wg.Done()
		for event := range ch {
			if err := auditor.Audit(event); err != nil {
				err = errors.Wrap(err, "audit event")
				logger.Log.Error(err.Error())
			}
		}
	}()
}

func (m *AuditManager) Notify(event model.AuditEvent) {
	select {
	case m.notifyCh <- event:
	default:
		logger.Log.Warn("audit channel full, dropping event")
	}
}

func (m *AuditManager) Close() {
	close(m.done)
	for _, w := range m.wrappers {
		close(w.ch)
	}
	m.wg.Wait()
	for _, w := range m.wrappers {
		if err := w.auditor.Close(); err != nil {
			err = errors.Wrap(err, "close auditor")
			logger.Log.Error(err.Error())
		}
	}
}
