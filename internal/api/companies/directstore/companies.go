package directstore

import (
	"context"

	"github.com/google/uuid"

	"github.com/RavisMsk/xmcompanies/internal/api/companies"
	"github.com/RavisMsk/xmcompanies/internal/api/models"
	storeModels "github.com/RavisMsk/xmcompanies/internal/companies/models"
	"github.com/RavisMsk/xmcompanies/internal/companies/store"
)

type Companies struct {
	store store.Store
}

func NewDirectStoreCompanies(store store.Store) *Companies {
	return &Companies{store}
}

func (c *Companies) Search(
	ctx context.Context,
	query companies.SearchFilters,
	skip uint64,
	limit uint64,
) ([]*models.Company, error) {
	results, err := c.store.Search(ctx, store.CompanyOptFields(query), skip, limit)
	if err != nil {
		return nil, err
	}
	companies := make([]*models.Company, len(results))
	for idx, result := range results {
		companies[idx] = &models.Company{
			ID:      result.ID,
			Name:    result.Name,
			Code:    result.Code,
			Country: result.Country,
			Website: result.Website,
			Phone:   result.Phone,
		}
	}
	return companies, nil
}

func (c *Companies) Get(ctx context.Context, id string) (*models.Company, error) {
	company, err := c.store.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &models.Company{
		ID:      company.ID,
		Name:    company.Name,
		Code:    company.Code,
		Country: company.Country,
		Website: company.Website,
		Phone:   company.Phone,
	}, nil
}

func (c *Companies) Create(ctx context.Context, company companies.CompanyFields) (string, error) {
	storeModel := storeModels.Company{
		ID:      uuid.New().String(),
		Name:    company.Name,
		Code:    company.Code,
		Country: company.Country,
		Website: company.Website,
		Phone:   company.Phone,
	}
	return storeModel.ID, c.store.Insert(ctx, &storeModel)
}

func (c *Companies) Update(ctx context.Context, id string, update companies.UpdateFields) error {
	return c.store.Update(ctx, id, store.CompanyOptFields(update))
}

func (c *Companies) Delete(ctx context.Context, id string) error {
	return c.store.Delete(ctx, id)
}
