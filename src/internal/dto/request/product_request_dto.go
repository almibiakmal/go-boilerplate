package requestdto

import (
	"strings"

	servicedto "maj-core-srvc/src/internal/dto/service"
)

// CreateProductRequest merepresentasikan payload JSON HTTP untuk membuat produk baru.
type CreateProductRequest struct {
	SKU   string `json:"sku" validate:"required,min=2,max=64"`
	Name  string `json:"name" validate:"required,min=2,max=150"`
	Price int64  `json:"price" validate:"gte=0"`
}

// ToServiceInput mengubah HTTP request payload menjadi data input untuk layer Service.
func (r CreateProductRequest) ToServiceInput() servicedto.CreateProductInput {
	return servicedto.CreateProductInput{
		SKU:   strings.TrimSpace(r.SKU),
		Name:  strings.TrimSpace(r.Name),
		Price: r.Price,
	}
}

// UpdateProductRequest merepresentasikan payload JSON HTTP untuk mengubah produk.
type UpdateProductRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=150"`
	Price int64  `json:"price" validate:"gte=0"`
}

// ToServiceInput mengubah HTTP request payload + path param ID menjadi data input untuk layer Service.
func (r UpdateProductRequest) ToServiceInput(id uint64) servicedto.UpdateProductInput {
	return servicedto.UpdateProductInput{
		ID:    id,
		Name:  strings.TrimSpace(r.Name),
		Price: r.Price,
	}
}

// ListProductRequest merepresentasikan query parameter HTTP untuk list produk.
type ListProductRequest struct {
	Page    int    `query:"page" validate:"omitempty,gte=1"`
	PerPage int    `query:"per_page" validate:"omitempty,gte=1,lte=100"`
	Search  string `query:"search" validate:"omitempty,max=100"`
}

// ToServiceFilter mengubah HTTP query params menjadi filter untuk layer Service.
func (r ListProductRequest) ToServiceFilter() servicedto.ListProductFilter {
	return servicedto.ListProductFilter{
		Page:    r.Page,
		PerPage: r.PerPage,
		Search:  strings.TrimSpace(r.Search),
	}
}
