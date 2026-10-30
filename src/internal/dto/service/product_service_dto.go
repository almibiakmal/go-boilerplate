package servicedto

// CreateProductInput adalah data bersih yang diterima oleh ProductService.Create
// (bisa berasal dari HTTP Handler, Event Consumer, maupun CLI).
type CreateProductInput struct {
	SKU   string `json:"sku" validate:"required,min=2,max=64"`
	Name  string `json:"name" validate:"required,min=2,max=150"`
	Price int64  `json:"price" validate:"gte=0"`
}

// UpdateProductInput adalah data bersih yang diterima oleh ProductService.Update.
type UpdateProductInput struct {
	ID    uint64 `json:"id" validate:"required,gt=0"`
	Name  string `json:"name" validate:"required,min=2,max=150"`
	Price int64  `json:"price" validate:"gte=0"`
}

// ListProductFilter adalah parameter filter & pagination untuk ProductService.List.
type ListProductFilter struct {
	Page    int
	PerPage int
	Search  string
}
