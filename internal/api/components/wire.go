//go:build wireinject
// +build wireinject

package components

import (
	"net/http"

	"github.com/google/wire"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/RavisMsk/xmcompanies/internal/api/api"
	"github.com/RavisMsk/xmcompanies/internal/api/companies"
	"github.com/RavisMsk/xmcompanies/internal/api/companies/directstore"
	"github.com/RavisMsk/xmcompanies/internal/api/ipchecker"
	companiesStore "github.com/RavisMsk/xmcompanies/internal/companies/store"
	mongoCompanies "github.com/RavisMsk/xmcompanies/internal/companies/store/mongo"
	"github.com/RavisMsk/xmcompanies/internal/pkg/ipapi"
)

func InitializeAssembly(cfgPath string) (*Assembly, error) {
	wire.Build(
		NewAssembly,
		ParseYAMLConfig,
		createLogger,
		createMongoClient,
		createMongoCompanies,
		createDirectMongoLayer,
		createAPI,
		createIPAPI,
		createIPChecker,
	)
	return &Assembly{}, nil
}

func createLogger(cfg *Config) *zap.Logger {
	lvl := zap.InfoLevel
	switch cfg.LogLevel {
	case "debug":
		lvl = zap.DebugLevel
	case "info":
		lvl = zap.InfoLevel
	case "warn":
		lvl = zap.WarnLevel
	case "error":
		lvl = zap.ErrorLevel
	}
	alvl := zap.NewAtomicLevelAt(lvl)
	zcfg := zap.NewProductionConfig()
	zcfg.Level = alvl
	zcfg.Encoding = "json"
	zcfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	zcfg.Sampling = nil
	logger, _ := zcfg.Build()
	return logger.Named("api")
}

func createMongoClient(cfg *Config, logger *zap.Logger) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.MongoURL))
	if err != nil {
		return nil, err
	}
	return client, nil
}

func createMongoCompanies(client *mongo.Client, logger *zap.Logger) companiesStore.Store {
	db := client.Database("xm")
	return mongoCompanies.NewStore(db.Collection("companies"))
}

func createDirectMongoLayer(store companiesStore.Store) companies.Companies {
	return directstore.NewDirectStoreCompanies(store)
}

func createAPI(
	cfg *Config,
	companies companies.Companies,
	ipChecker ipchecker.Checker,
	logger *zap.Logger,
) *api.API {
	return api.NewAPI(cfg, companies, ipChecker, logger.Named("api"))
}

func createIPAPI(cfg *Config) *ipapi.Client {
	return ipapi.NewClient(cfg.IPAPIKey, http.DefaultClient)
}

func createIPChecker(client *ipapi.Client) ipchecker.Checker {
	return ipchecker.NewIPAPIChecker(client)
}
