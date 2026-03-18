package domain

type ImageUploadResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"isPrimary"`
}

type ReorderRequest struct {
	ImageIDs []string `json:"imageIds"`
}
