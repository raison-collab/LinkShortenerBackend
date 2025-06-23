package dto

// ErrorResponse представляет стандартный ответ об ошибке
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}

// PaginationRequest представляет параметры пагинации
type PaginationRequest struct {
	Page  int `form:"page,default=1" binding:"min=1"`
	Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}

// GetOffset вычисляет offset для пагинации
func (p *PaginationRequest) GetOffset() int {
	return (p.Page - 1) * p.Limit
}
