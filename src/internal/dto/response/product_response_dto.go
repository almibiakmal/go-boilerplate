package responsedto

import (
	"time"

	"maj-core-srvc/src/internal/model"
)

// ProductResponse adalah struktur JSON yang dikembalikan ke client HTTP.
type ProductResponse struct {
	ID        uint64    `json:"id"`
	SKU       string    `json:"sku"`
	Name      string    `json:"name"`
	Price     int64     `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToProductResponse memetakan entity model.Product ke *ProductResponse.
func ToProductResponse(m *model.Product) *ProductResponse {
	if m == nil {
		return nil
	}
	return &ProductResponse{
		ID:        m.ID,
		SKU:       m.SKU,
		Name:      m.Name,
		Price:     m.Price,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

// ToProductListResponse memetakan slice model.Product ke slice ProductResponse tanpa alokasi pointer per-elemen.
func ToProductListResponse(items []model.Product) []ProductResponse {
	res := make([]ProductResponse, len(items))
	for i := range items {
		res[i] = ProductResponse{
			ID:        items[i].ID,
			SKU:       items[i].SKU,
			Name:      items[i].Name,
			Price:     items[i].Price,
			CreatedAt: items[i].CreatedAt,
			UpdatedAt: items[i].UpdatedAt,
		}
	}
	return res
}
