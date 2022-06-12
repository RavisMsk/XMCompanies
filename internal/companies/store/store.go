package store

import (
	"context"
	"errors"

	"github.com/RavisMsk/xmcompanies/internal/companies/models"
)

var (
	ErrNotFound = errors.New("not found")
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

type Store interface {
	Get(ctx context.Context, id string) (*models.Company, error)
	Insert(ctx context.Context, company *models.Company) error
	Update(
		ctx context.Context,
		id string,
		fields CompanyOptFields,
	) error
	Delete(ctx context.Context, id string) error
	Search(
		ctx context.Context,
		query CompanyOptFields,
		skip,
		limit uint64,
	) ([]*models.Company, error)
}
