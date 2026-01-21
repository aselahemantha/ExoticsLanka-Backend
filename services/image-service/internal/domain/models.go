package domain

type ImageUploadResponse struct {
	ID        string `json:"id"`
	URL       string `json:"url"`
	IsPrimary bool   `json:"is_primary"`
}

type ReorderRequest struct {
	ImageIDs []string `json:"image_ids"`
}
