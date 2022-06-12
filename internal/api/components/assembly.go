package components

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"

	"github.com/RavisMsk/xmcompanies/internal/api/api"
)

type Assembly struct {
	Log *zap.Logger

	config *Config
	mongo  *mongo.Client
	api    *api.API
}

func NewAssembly(
	cfg *Config,
	mongo *mongo.Client,
	api *api.API,
	log *zap.Logger,
) *Assembly {
	return &Assembly{log, cfg, mongo, api}
}

func (a *Assembly) Run() {
	a.Log.Info("starting up api")

	a.Log.Info("connecting to mongo")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.mongo.Connect(ctx); err != nil {
		a.Log.Fatal("error connecting to mongo", zap.Error(err))
	}
	if err := a.mongo.Ping(ctx, readpref.Primary()); err != nil {
		a.Log.Fatal("error pinging mongo primary", zap.Error(err))
	}

	if err := a.api.Run(); err != nil {
		a.Log.Fatal("error starting API", zap.Error(err))
	}

	a.Log.Info("api started")
}

func (a *Assembly) Stop() error {
	a.Log.Warn("stopping api")
	a.api.Stop()
	a.Log.Warn("api stopped")
	return nil
}
