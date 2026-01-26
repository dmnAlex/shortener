package audit

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/dmnAlex/shortener/internal/model"
	"github.com/pkg/errors"
)

type RemoteAuditor struct {
	url    string
	client *http.Client
}

func NewRemoteAuditor(url string) *RemoteAuditor {
	return &RemoteAuditor{
		url:    url,
		client: &http.Client{},
	}
}

func (a *RemoteAuditor) Audit(event model.AuditEvent) error {
	data, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "marshal event")
	}

	res, err := a.client.Post(a.url, "application/json", bytes.NewReader(data))
	if err != nil {
		return errors.Wrap(err, "do post request")
	}
	defer res.Body.Close()

	if res.StatusCode/2 != 2 {
		return errors.Errorf("remote audit failed with status %d", res.StatusCode)
	}

	return nil
}

func (a *RemoteAuditor) Close() error {
	return nil
}
