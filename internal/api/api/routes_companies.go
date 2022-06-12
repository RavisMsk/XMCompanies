package api

import (
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RavisMsk/xmcompanies/internal/api/companies"
	"github.com/RavisMsk/xmcompanies/internal/pkg/countries"
)

func (a *API) handleListCompanies(c *gin.Context, log *zap.Logger) {
	nameFilter := c.Query("name")
	codeFilter := c.Query("code")
	countryFilter := c.Query("country")
	websiteFilter := c.Query("website")
	phoneFilter := c.Query("phone")

	var (
		page  uint64
		limit uint64 = 20
		err   error
	)
	pageString := c.Query("cursor")
	if len(pageString) > 0 {
		page, err = strconv.ParseUint(pageString, 10, 64)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
	}
	limitString := c.Query("limit")
	if len(limitString) > 0 {
		limit, err = strconv.ParseUint(limitString, 10, 64)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		if limit < 2 {
			c.Status(http.StatusBadRequest)
			return
		}
	}

	query := companies.SearchFilters{}
	if len(nameFilter) > 0 {
		query.Name = &nameFilter
	}
	if len(codeFilter) > 0 {
		query.Code = &codeFilter
	}
	if len(countryFilter) > 0 {
		query.Country = &countryFilter
	}
	if len(websiteFilter) > 0 {
		query.Website = &websiteFilter
	}
	if len(phoneFilter) > 0 {
		query.Phone = &phoneFilter
	}
	companies, err := a.companies.Search(getCtx(c), query, page, limit)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Error("companies search error", zap.Error(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": companies,
	})
}

func (a *API) handleGetCompany(c *gin.Context, log *zap.Logger) {
	companyID := c.Param("companyID")
	log.Info("fetching company", zap.String("id", companyID))
	company, err := a.companies.Get(getCtx(c), companyID)
	if err == companies.ErrNotFound {
		c.Status(http.StatusNotFound)
		log.Error("company not found", zap.String("id", companyID))
		return
	} else if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Error("unexpected error fetching company", zap.String("id", companyID), zap.Error(err))
		return
	}
	c.JSON(http.StatusOK, company)
}

type createCompanyRequest struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Country string `json:"country"`
	Website string `json:"website"`
	Phone   string `json:"phone"`
}

func (a *API) handleCreateCompany(c *gin.Context, log *zap.Logger) {
	var request createCompanyRequest
	if err := c.BindJSON(&request); err != nil {
		c.Status(http.StatusBadRequest)
		log.Error("couldnt unmarshal create company request", zap.Error(err))
		return
	}

	log.Info(
		"create company request",
		zap.String("name", request.Name),
		zap.String("code", request.Code),
		zap.String("country", request.Country),
		zap.String("website", request.Website),
		zap.String("phone", request.Phone),
	)

	var (
		errs   []error
		fields companies.CompanyFields
	)

	if processedName, err := validatedName(request.Name); err != nil {
		errs = append(errs, err)
	} else {
		fields.Name = processedName
	}

	if processedCode, err := validatedCode(request.Code); err != nil {
		errs = append(errs, err)
	} else {
		fields.Code = processedCode
	}

	if !validCountry(request.Country) {
		errs = append(errs, errors.New("invalid country"))
	} else {
		fields.Country = request.Country
	}

	if processedWebsite, err := validatedWebsite(request.Website); err != nil {
		errs = append(errs, err)
	} else {
		fields.Website = processedWebsite
	}

	if normalizedPhone, err := normalizePhone(request.Phone); err != nil {
		errs = append(errs, err)
	} else {
		fields.Phone = normalizedPhone
	}

	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": errs,
		})
		return
	}

	companyID, err := a.companies.Create(getCtx(c), fields)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Error("error creating company", zap.Error(err))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id": companyID,
	})
}

type companyUpdateRequest struct {
	Name    *string `json:"name"`
	Code    *string `json:"code"`
	Country *string `json:"country"`
	Website *string `json:"website"`
	Phone   *string `json:"phone"`
}

func (a *API) handleUpdateCompany(c *gin.Context, log *zap.Logger) {
	var request companyUpdateRequest
	if err := c.BindJSON(&request); err != nil {
		c.Status(http.StatusBadRequest)
		log.Error("couldnt unmarshal company request", zap.Error(err))
		return
	}

	companyID := c.Param("companyID")
	var errs []error
	update := companies.UpdateFields{}

	if request.Name != nil {
		if processedName, err := validatedName(*request.Name); err != nil {
			errs = append(errs, err)
		} else {
			update.Name = &processedName
		}
	}

	if request.Code != nil {
		if processedCode, err := validatedCode(*request.Code); err != nil {
			errs = append(errs, err)
		} else {
			update.Code = &processedCode
		}
	}

	if request.Country != nil {
		if validCountry(*request.Country) {
			update.Country = request.Country
		} else {
			errs = append(errs, errors.New("invalid country"))
		}
	}

	if request.Website != nil {
		if processedWebsite, err := validatedWebsite(*request.Website); err != nil {
			errs = append(errs, err)
		} else {
			update.Website = &processedWebsite
		}
	}

	if request.Phone != nil {
		if normalizedPhone, err := normalizePhone(*request.Phone); err != nil {
			errs = append(errs, err)
		} else {
			update.Phone = &normalizedPhone
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"errors": errs,
		})
		return
	}

	if err := a.companies.Update(getCtx(c), companyID, update); err != nil {
		c.Status(http.StatusInternalServerError)
		log.Error("error updating company", zap.Error(err))
		return
	}
	c.Status(http.StatusOK)
}

func (a *API) handleDeleteCompany(c *gin.Context, log *zap.Logger) {
	companyID := c.Param("companyID")
	err := a.companies.Delete(getCtx(c), companyID)
	if err == companies.ErrNotFound {
		c.Status(http.StatusNotFound)
		log.Error("company to delete not found", zap.String("id", companyID))
		return
	} else if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Error("error deleting company", zap.String("id", companyID), zap.Error(err))
		return
	}
	c.Status(http.StatusOK)
}

var nameMatcher = regexp.MustCompile("^[A-Za-z ]+$").MatchString

func validatedName(name string) (string, error) {
	name = strings.Trim(name, " \n")
	if len(name) < 4 {
		return "", errors.New("company name must be at least 4 characters")
	}
	if !nameMatcher(name) {
		return "", errors.New("company name can contain only letters and spaces")
	}
	return name, nil
}

var codeMatcher = regexp.MustCompile("^[A-Z]+$").MatchString

func validatedCode(code string) (string, error) {
	code = strings.Trim(code, " \n")
	if len(code) < 2 {
		return "", errors.New("")
	}
	if !codeMatcher(code) {
		return "", errors.New("company code can contain only uppercase letters")
	}
	return code, nil
}

func validCountry(country string) bool {
	return countries.IsValidCountry(country)
}

func validatedWebsite(website string) (string, error) {
	_, err := url.ParseRequestURI(website)
	if err != nil {
		return "", errors.New("website is not valid url")
	}
	return website, nil
}

func normalizePhone(phone string) (string, error) {
	return phone, nil
}
