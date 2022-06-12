package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/RavisMsk/xmcompanies/internal/companies/models"
	"github.com/RavisMsk/xmcompanies/internal/companies/store"
)

type Store struct {
	col *mongo.Collection
}

func NewStore(col *mongo.Collection) *Store {
	return &Store{col}
}

func (s *Store) Get(ctx context.Context, id string) (*models.Company, error) {
	query := bson.M{
		"id": id,
	}
	result := s.col.FindOne(ctx, query)
	err := result.Err()
	if err == mongo.ErrNoDocuments {
		return nil, store.ErrNotFound
	} else if err != nil {
		return nil, err
	}

	var company models.Company
	if err = result.Decode(&company); err != nil {
		return nil, err
	}

	return &company, nil
}

func (s *Store) Insert(ctx context.Context, company *models.Company) error {
	company.CreatedAt = time.Now()
	company.UpdatedAt = nil
	_, err := s.col.InsertOne(ctx, company)
	return err
}

func (s *Store) Update(
	ctx context.Context,
	id string,
	fields store.CompanyOptFields,
) error {
	query := bson.M{
		"id": id,
	}
	patch := bson.M{
		"updated_at": time.Now(),
	}
	if fields.Name != nil {
		patch["name"] = *fields.Name
	}
	if fields.Code != nil {
		patch["code"] = *fields.Code
	}
	if fields.Country != nil {
		patch["country"] = *fields.Country
	}
	if fields.Phone != nil {
		patch["phone"] = *fields.Phone
	}
	if fields.Website != nil {
		patch["website"] = *fields.Website
	}

	result, err := s.col.UpdateOne(ctx, query, bson.M{
		"$set": patch,
	})
	if err != nil {
		return err
	}
	if result.ModifiedCount < 1 {
		return store.ErrNotFound
	}
	return nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	query := bson.M{
		"id": id,
	}
	result, err := s.col.DeleteOne(ctx, query)
	if err != nil {
		return err
	}
	if result.DeletedCount < 1 {
		return store.ErrNotFound
	}
	return nil
}

func (s *Store) Search(
	ctx context.Context,
	query store.CompanyOptFields,
	skip, limit uint64,
) ([]*models.Company, error) {
	filter := bson.M{}
	if query.Name != nil {
		filter["name"] = query.Name
	}
	if query.Code != nil {
		filter["code"] = query.Code
	}
	if query.Country != nil {
		filter["country"] = query.Country
	}
	if query.Phone != nil {
		filter["phone"] = query.Phone
	}
	if query.Website != nil {
		filter["website"] = query.Website
	}

	cursor, err := s.col.Find(
		ctx,
		query,
		options.Find().SetSkip(int64(skip)).SetLimit(int64(limit)),
	)
	if err != nil {
		return nil, err
	}

	var results []*models.Company
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
