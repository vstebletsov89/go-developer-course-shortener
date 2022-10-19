// Package types contains set of structures for requests and response.
package types

// RequestJSON represents a link for json requests.
type RequestJSON struct {
	URL string `json:"url"`
}

// ResponseJSON represents a link for json responses.
type ResponseJSON struct {
	Result string `json:"result"`
}

// RequestBatchJSON represents a link for batch json requests.
type RequestBatchJSON struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ResponseBatchJSON represents a link for batch json responses.
type ResponseBatchJSON struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// RequestBatch represents a slice of links for batch json requests.
type RequestBatch []RequestBatchJSON

// ResponseBatch represents a slice of links for batch json responses.
type ResponseBatch []ResponseBatchJSON
