package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/RavisMsk/xmcompanies/internal/api/companies"
	"github.com/RavisMsk/xmcompanies/internal/api/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const allowedTestCountry = "Test"

type testConfig struct{}

func (c *testConfig) GetDebug() bool                    { return true }
func (c *testConfig) GetListenAddr() string             { return "" }
func (c *testConfig) GetTimeoutDuration() time.Duration { return 10 * time.Second }
func (c *testConfig) GetAllowedCountries() []string     { return []string{allowedTestCountry} }

func createTestAPI(
	companies *companiesLayerMock,
	ipChecker *ipCheckerMock,
) *API {
	log, _ := zap.NewDevelopment()
	var cfg testConfig
	return NewAPI(&cfg, companies, ipChecker, log)
}

type companiesLayerMock struct {
	mock.Mock
}

func (m *companiesLayerMock) Search(
	ctx context.Context,
	query companies.SearchFilters,
	skip,
	limit uint64,
) ([]*models.Company, error) {
	args := m.Called(query, skip, limit)
	return args.Get(0).([]*models.Company), args.Error(1)
}
func (m *companiesLayerMock) Get(ctx context.Context, id string) (*models.Company, error) {
	args := m.Called(id)
	var company *models.Company
	if args.Get(0) != nil {
		company = args.Get(0).(*models.Company)
	}
	return company, args.Error(1)
}
func (m *companiesLayerMock) Create(ctx context.Context, fields companies.CompanyFields) (string, error) {
	args := m.Called(fields)
	return args.String(0), args.Error(1)
}
func (m *companiesLayerMock) Update(ctx context.Context, id string, update companies.UpdateFields) error {
	args := m.Called(id, update)
	return args.Error(0)
}
func (m *companiesLayerMock) Delete(ctx context.Context, id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type ipCheckerMock struct {
	mock.Mock
}

func (m *ipCheckerMock) GetIPCountry(ip string) (string, error) {
	args := m.Called(ip)
	return args.String(0), args.Error(1)
}

func TestModifyCompanyFromForbiddenAddress(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		comps := &companiesLayerMock{}
		checker := &ipCheckerMock{}
		checker.On("GetIPCountry", "44.44.44.44").Return("Unwhitelisted", nil)

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/companies", nil)
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
		checker.AssertExpectations(t)
	})
	t.Run("delete", func(t *testing.T) {
		comps := &companiesLayerMock{}
		checker := &ipCheckerMock{}
		checker.On("GetIPCountry", "44.44.44.44").Return("Unwhitelisted", nil)

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/companies/1234", nil)
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
		checker.AssertExpectations(t)
	})
}

var validCompany = models.Company{
	ID:      "1234",
	Name:    "Valid Name",
	Code:    "VN",
	Country: "Cyprus",
	Website: "http://company.valid/",
	Phone:   "79991234567",
}

func TestCreateCompany(t *testing.T) {
	cases := []struct {
		name         string
		body         gin.H
		expectedCode int
		expectedId   string
		errorsCnt    int
	}{
		{
			name: "valid company",
			body: gin.H{
				"name":    "Valid Name",
				"code":    "VN",
				"country": "Cyprus",
				"website": "http://valid.name/",
				"phone":   "79991234567",
			},
			expectedCode: http.StatusCreated,
			expectedId:   "1234",
			errorsCnt:    0,
		},
		{
			name: "invalid empty name",
			body: gin.H{
				"name":    "",
				"code":    "VN",
				"country": "Cyprus",
				"website": "http://valid.name/",
				"phone":   "79991234567",
			},
			expectedCode: http.StatusBadRequest,
			errorsCnt:    1,
		},
		{
			name: "invalid symbols name",
			body: gin.H{
				"name":    "345678-12asd",
				"code":    "VN",
				"country": "Cyprus",
				"website": "http://valid.name/",
				"phone":   "79991234567",
			},
			expectedCode: http.StatusBadRequest,
			errorsCnt:    1,
		},
		{
			name: "invalid symbols code",
			body: gin.H{
				"name":    "Valid Name",
				"code":    "abcd",
				"country": "Cyprus",
				"website": "http://valid.name/",
				"phone":   "79991234567",
			},
			expectedCode: http.StatusBadRequest,
			errorsCnt:    1,
		},
		{
			name: "invalid length code",
			body: gin.H{
				"name":    "Valid Name",
				"code":    "A",
				"country": "Cyprus",
				"website": "http://valid.name/",
				"phone":   "79991234567",
			},
			expectedCode: http.StatusBadRequest,
			errorsCnt:    1,
		},
		{
			name: "invalid country",
			body: gin.H{
				"name":    "Valid Name",
				"code":    "VN",
				"country": "atlantis",
				"website": "http://valid.name/",
				"phone":   "79991234567",
			},
			expectedCode: http.StatusBadRequest,
			errorsCnt:    1,
		},
		{
			name: "invalid website",
			body: gin.H{
				"name":    "Valid Name",
				"code":    "VN",
				"country": "Cyprus",
				"website": "here-should-be-a-link",
				"phone":   "79991234567",
			},
			expectedCode: http.StatusBadRequest,
			errorsCnt:    1,
		},
		{
			name: "multiple errors",
			body: gin.H{
				"name":    "invalid1234name",
				"code":    "a",
				"country": "atlantis",
				"website": "http://valid.name/",
				"phone":   "79991234567",
			},
			expectedCode: http.StatusBadRequest,
			errorsCnt:    3,
		},
	}

	for _, cs := range cases {
		t.Run(cs.name, func(t *testing.T) {

			comps := &companiesLayerMock{}
			if cs.expectedCode == http.StatusCreated {
				comps.On("Create", companies.CompanyFields{
					Name:    cs.body["name"].(string),
					Code:    cs.body["code"].(string),
					Country: cs.body["country"].(string),
					Website: cs.body["website"].(string),
					Phone:   cs.body["phone"].(string),
				}).Return(cs.expectedId, nil)
			}

			checker := &ipCheckerMock{}
			checker.On("GetIPCountry", "44.44.44.44").Return(allowedTestCountry, nil)

			api := createTestAPI(comps, checker)
			engine := api.createEngine()

			bodyBytes, _ := json.Marshal(cs.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/v1/companies", bytes.NewReader(bodyBytes))
			req.RemoteAddr = "44.44.44.44:54321"
			engine.ServeHTTP(w, req)

			assert.Equal(t, cs.expectedCode, w.Code)
			if cs.expectedCode == http.StatusCreated {
				var response gin.H
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, cs.expectedId, response["id"])
				assert.Equal(t, 1, len(response))
			} else if cs.expectedCode == http.StatusBadRequest {
				var response gin.H
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, cs.errorsCnt, len(response["errors"].([]interface{})))
			}

			checker.AssertExpectations(t)

		})
	}

	t.Run("other body format request", func(t *testing.T) {
		comps := &companiesLayerMock{}
		checker := &ipCheckerMock{}
		checker.On("GetIPCountry", "44.44.44.44").Return(allowedTestCountry, nil)

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		bodyBytes := []byte("name=ValidName&code=VN&country=Cyprus&website=http://company.valid/&phone=79991234567")
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/companies", bytes.NewReader(bodyBytes))
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		checker.AssertExpectations(t)
	})
}

func TestGetCompany(t *testing.T) {
	t.Run("existing company", func(t *testing.T) {
		comps := &companiesLayerMock{}
		comps.On("Get", "1234").Return(&validCompany, nil)

		checker := &ipCheckerMock{}

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/companies/1234", nil)
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response gin.H
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, validCompany.ID, response["id"])
		assert.Equal(t, validCompany.Name, response["name"])
		assert.Equal(t, validCompany.Code, response["code"])
		assert.Equal(t, validCompany.Country, response["country"])
		assert.Equal(t, validCompany.Website, response["website"])
		assert.Equal(t, validCompany.Phone, response["phone"])

		checker.AssertExpectations(t)
	})

	t.Run("non-existent company", func(t *testing.T) {
		comps := &companiesLayerMock{}
		comps.On("Get", "1234").Return(nil, companies.ErrNotFound)

		checker := &ipCheckerMock{}

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/companies/1234", nil)
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
		checker.AssertExpectations(t)
	})

	t.Run("unexpected get error", func(t *testing.T) {
		comps := &companiesLayerMock{}
		comps.On("Get", "1234").Return(nil, errors.New("unexpected error"))

		checker := &ipCheckerMock{}

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/companies/1234", nil)
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		checker.AssertExpectations(t)
	})
}

func TestDeleteCompany(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		comps := &companiesLayerMock{}
		comps.On("Delete", "1234").Return(nil)

		checker := &ipCheckerMock{}
		checker.On("GetIPCountry", "44.44.44.44").Return(allowedTestCountry, nil)

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/companies/1234", nil)
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		checker.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		comps := &companiesLayerMock{}
		comps.On("Delete", "1234").Return(companies.ErrNotFound)

		checker := &ipCheckerMock{}
		checker.On("GetIPCountry", "44.44.44.44").Return(allowedTestCountry, nil)

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/companies/1234", nil)
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		checker.AssertExpectations(t)
	})

	t.Run("unexpected error", func(t *testing.T) {
		comps := &companiesLayerMock{}
		comps.On("Delete", "1234").Return(errors.New("unexpected error"))

		checker := &ipCheckerMock{}
		checker.On("GetIPCountry", "44.44.44.44").Return(allowedTestCountry, nil)

		api := createTestAPI(comps, checker)
		engine := api.createEngine()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/v1/companies/1234", nil)
		req.RemoteAddr = "44.44.44.44:54321"
		engine.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		checker.AssertExpectations(t)
	})
}
