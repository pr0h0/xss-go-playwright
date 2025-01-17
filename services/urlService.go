package services

import (
	"strings"
	"xss/utils"

	urlutil "github.com/projectdiscovery/utils/url"
)

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

func (us *UrlService) CombineUrlQueryWithPayload(url string, payloads []string) []string {
	combinedUrls := []string{}

	if !us.ValidateUrl(url) {
		utils.Log.Error("Invalid URL:", url)
		return combinedUrls
	}

	if strings.Contains(url, "{payload}") {
		for _, payload := range payloads {
			combinedUrls = append(combinedUrls, strings.ReplaceAll(url, "{payload}", payload))
		}
		return combinedUrls
	} else {
		parsedUrl, err := urlutil.Parse(url)
		if err != nil {
			return combinedUrls
		}
		if parsedUrl.RawQuery != "" {
			parsedUrl.Query().Iterate(func(key string, value []string) bool {
				for _, payload := range payloads {
					clonedUrl, _ := urlutil.Parse(url)
					clonedUrl.Query().Set(key, payload)
					combinedUrls = append(combinedUrls, clonedUrl.String())
				}
				return true
			})
		}
	}

	return combinedUrls
}
