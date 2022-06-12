package api

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/RavisMsk/xmcompanies/internal/api/companies"
	"github.com/RavisMsk/xmcompanies/internal/api/ipchecker"
	"github.com/RavisMsk/xmcompanies/internal/pkg/structs"
)

type Config interface {
	GetDebug() bool
	GetListenAddr() string
	GetTimeoutDuration() time.Duration
	GetAllowedCountries() []string
}

type API struct {
	cfg       Config
	companies companies.Companies
	ipChecker ipchecker.Checker
	log       *zap.Logger

	stopping int32
	wg       sync.WaitGroup
}

func NewAPI(
	cfg Config,
	companies companies.Companies,
	ipChecker ipchecker.Checker,
	log *zap.Logger,
) *API {
	return &API{
		cfg:       cfg,
		companies: companies,
		ipChecker: ipChecker,
		log:       log,
	}
}

func (a *API) Run() error {
	if !a.cfg.GetDebug() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := a.createEngine()
	go func() {
		if err := r.Run(a.cfg.GetListenAddr()); err != nil {
			a.log.Fatal("error running api", zap.Error(err))
		}
	}()

	return nil
}

func (a *API) Stop() {
	atomic.AddInt32(&a.stopping, 1)
	a.wg.Wait()
}

func (a *API) createEngine() *gin.Engine {
	r := gin.New()
	r.Use(
		ginzap.Ginzap(a.log, time.RFC3339, true),
		ginzap.RecoveryWithZap(a.log, true),
	)

	r.Use(a.reqSetupMiddleware)
	r.Use(a.shutdownMiddleware)

	r.GET("/v1", func(c *gin.Context) {
		c.Status(200)
	})

	allowedCountries := structs.NewStringSet()
	allowedCountries.Add(a.cfg.GetAllowedCountries()...)

	v1 := r.Group("/v1")
	v1.GET("/companies", a.wrapHandler(a.handleListCompanies))
	v1.GET("/companies/:companyID", a.wrapHandler(a.handleGetCompany))
	v1.PUT("/companies/:companyID", a.wrapHandler(a.handleUpdateCompany))
	v1.POST(
		"/companies",
		IPCheckingMiddleware(a.ipChecker, *allowedCountries),
		a.wrapHandler(a.handleCreateCompany),
	)
	v1.DELETE(
		"/companies/:companyID",
		IPCheckingMiddleware(a.ipChecker, *allowedCountries),
		a.wrapHandler(a.handleDeleteCompany),
	)

	return r
}

func (a *API) reqSetupMiddleware(c *gin.Context) {
	reqID := uuid.New().String()
	reqLogger := a.log.With(
		zap.String("path", c.Request.URL.Path),
		zap.String("reqID", reqID),
	)

	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.GetTimeoutDuration())
	defer cancel()

	setReqID(c, reqID)
	setLogger(c, reqLogger)
	setCtx(c, ctx)

	c.Next()
}

func (a *API) wrapHandler(f func(*gin.Context, *zap.Logger)) func(*gin.Context) {
	return func(c *gin.Context) {
		f(c, getLogger(c))
	}
}

func (a *API) shutdownMiddleware(c *gin.Context) {
	if atomic.LoadInt32(&a.stopping) > 0 {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	a.wg.Add(1)
	defer a.wg.Done()
	c.Next()
}
