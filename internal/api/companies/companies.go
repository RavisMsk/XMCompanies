package companies

import (
	"context"
	"errors"

	"github.com/RavisMsk/xmcompanies/internal/api/models"
)

type CompanyFields struct {
	Name    string
	Code    string
	Country string
	Website string
	Phone   string
}

type CompanyOptFields struct {
	Name    *string
	Code    *string
	Country *string
	Website *string
	Phone   *string
}

type SearchFilters CompanyOptFields

type UpdateFields CompanyOptFields

var (
	ErrNotFound = errors.New("not found")
)

type Companies interface {
	Search(
		ctx context.Context,
		query SearchFilters,
		skip,
		limit uint64,
	) ([]*models.Company, error)
	Get(ctx context.Context, id string) (*models.Company, error)
	Create(ctx context.Context, fields CompanyFields) (string, error)
	Update(ctx context.Context, id string, update UpdateFields) error
	Delete(ctx context.Context, id string) error
}
