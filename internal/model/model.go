package model

type URLRecord struct {
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}

type ShortenRequest struct {
	URL string `json:"url" binding:"required"`
}

type ShortenResponse struct {
	Result string `json:"result"`
}

type ShortenBatchRequest struct {
	CorrelationID string `json:"correlation_id" binding:"required"`
	OriginalURL   string `json:"original_url" binding:"required"`
}

type ShortenBatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type UserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (m *UserURLsResponse) AsIfaceList() []any {
	return []any{&m.ShortURL, &m.OriginalURL}
}

type Caller struct {
	UserID string
}

type DeleteTask struct {
	UserID  string
	ShortID string
}
