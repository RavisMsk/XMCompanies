package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/RavisMsk/xmcompanies/internal/api/ipchecker"
	"github.com/RavisMsk/xmcompanies/internal/pkg/structs"
)

func IPCheckingMiddleware(checker ipchecker.Checker, allowedCountries structs.StringSet) func(*gin.Context) {
	return func(c *gin.Context) {
		log := getLogger(c)
		clientIP := c.ClientIP()
		clientCountry, err := checker.GetIPCountry(clientIP)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error(
				"error fetching client ip country",
				zap.String("ip", clientIP),
				zap.Error(err),
			)
			return
		}
		if !allowedCountries.Has(clientCountry) {
			c.AbortWithStatus(http.StatusForbidden)
			log.Error(
				"nonwhitelisted client country",
				zap.String("ip", clientIP),
				zap.String("country", clientCountry),
			)
			return
		}
		log.Info(
			"validated client call country",
			zap.String("ip", clientIP),
			zap.String("country", clientCountry),
		)
		c.Next()
	}
}
