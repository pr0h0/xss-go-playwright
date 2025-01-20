package services

import (
	"fmt"
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

func (us *UrlService) CombineRequestWithPayload(request map[string]string, payloads string) (map[string]string, error) {
	val, ok := request["_ BODY"]
	if !ok {
		return nil, fmt.Errorf("request body not found")
	}

	if strings.Contains(val, "{payload}") {
		request["_ BODY"] = strings.ReplaceAll(val, "{payload}", payloads)
		return request, nil
	}

	return nil, fmt.Errorf("payload not found in request body")
}

func (us *UrlService) ParseRequest(fileContent string) (request map[string]string, _ error) {
	request = make(map[string]string)
	utils.Log.Info("Parsing request file")
	fileContent = strings.Trim(fileContent, " ")
	fileContent = strings.Trim(fileContent, "\n")
	lines := strings.Split(fileContent, "\n")

	emptyLineCount := 0
	totalLinesParsed := 0
	for ix, line := range lines {
		line := strings.Trim(line, " ")

		if len(line) == 0 {
			emptyLineCount++
			continue
		}
		totalLinesParsed++

		// parse request line POST /path HTTP/1.1
		if ix == 0 {
			lineArr := strings.Split(strings.Trim(line, " "), " ")
			if len(lineArr) != 3 {
				utils.Log.Error("Invalid request format")
				return nil, fmt.Errorf("invalid request format")
			}
			request["_ METHOD"] = lineArr[0]
			request["_ PATH"] = lineArr[1]
			request["_ PROTOCOL"] = lineArr[2]

			continue
		}

		// if we have 2 empty lines already, (one after request line and one after headers), or if we are at the end of the file
		if emptyLineCount == 2 || (ix == len(lines)-1) {
			request["_ BODY"] = strings.Join(lines[ix:], "\n")
			break
		}

		// parse headers
		headerLine := strings.SplitN(line, ":", 2)
		if len(headerLine) != 2 {
			utils.Log.Error("Invalid header format", headerLine)
			return nil, fmt.Errorf("invalid header format")
		}

		request[headerLine[0]] = strings.Trim(headerLine[1], " ")
	}

	return request, nil
}
