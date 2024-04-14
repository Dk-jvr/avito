package Components

type (
	Banner struct {
		BannerId        int    `json:"banner_id"`
		TagIds          []int  `json:"tag_ids" validate:"required"`
		FeatureId       int    `json:"feature_id" validate:"required"`
		Content         string `json:"content" validate:"required"`
		IsActive        bool   `json:"is_active" validate:"required"`
		Version         int    `json:"version"`
		CreatedAtString string `json:"created_at"`
		UpdatedAtString string `json:"updated_at"`
	}
	ShortBanner struct {
		BannerId int    `json:"banner_id"`
		Content  string `json:"content"`
	}
)
