package services

import urlutil "github.com/projectdiscovery/utils/url"

type UrlService struct{}

var urlServiceInstance *UrlService = nil

// Singleton instance of UrlService
func GetUrlService() (*UrlService, error) {
	if urlServiceInstance == nil {
		urlServiceInstance = &UrlService{}
	}

	return urlServiceInstance, nil
}

// Validate URL
func (us *UrlService) ValidateUrl(url string) bool {
	if _, err := urlutil.Parse(url); err != nil {
		return false
	}
	return true
}
