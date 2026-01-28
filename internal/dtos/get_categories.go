package dtos

import "github.com/AlexMickh/shop-backend/internal/models"

type GetCategoriesResponse struct {
	Categories []category `json:"categories"`
}

type category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func ToGetCategoriesResponse(categories []models.Category) GetCategoriesResponse {
	resp := make([]category, 0, len(categories))

	for _, v := range categories {
		resp = append(resp, category{
			ID:   v.ID,
			Name: v.Name,
		})
	}

	return GetCategoriesResponse{
		Categories: resp,
	}
}
