package ipchecker

import "github.com/RavisMsk/xmcompanies/internal/pkg/ipapi"

type IPAPIChecker struct {
	client *ipapi.Client
}

func NewIPAPIChecker(client *ipapi.Client) *IPAPIChecker {
	return &IPAPIChecker{client}
}

func (c *IPAPIChecker) GetIPCountry(ip string) (string, error) {
	result, err := c.client.LookupIP(ip)
	if err != nil {
		return "", err
	}
	return result.CountryName, nil
}
