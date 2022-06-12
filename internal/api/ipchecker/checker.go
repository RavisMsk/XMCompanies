package ipchecker

type Checker interface {
	GetIPCountry(ip string) (string, error)
}
