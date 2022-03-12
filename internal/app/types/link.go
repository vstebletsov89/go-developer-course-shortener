package types

type Link struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type BatchLink struct {
	CorrelationID string
	ShortURL      string
	OriginalURL   string
}

type BatchLinks []BatchLink
