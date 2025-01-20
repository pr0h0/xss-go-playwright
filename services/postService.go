package services

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
	"xss/utils"

	"github.com/playwright-community/playwright-go"
)

type PostService struct {
	CombinedUrls []map[string]string
	Request      map[string]string
	Payloads     []string
	UUID         string

	fileService    *FileService
	argService     *ArgsService
	jsonService    *JsonService
	urlService     *UrlService
	requestService *RequestService
}

var postServiceInstance *PostService = nil

// Singleton instance of PostService
func GetPostService() (*PostService, error) {
	if postServiceInstance == nil {
		var fileService *FileService
		var argService *ArgsService
		var jsonService *JsonService
		var urlService *UrlService
		var requestService *RequestService
		var err error

		if fileService, err = GetFileService(); err != nil {
			return nil, err
		}

		if argService, err = GetArgsService(); err != nil {
			return nil, err
		}

		if jsonService, err = GetJsonService(); err != nil {
			return nil, err
		}

		if urlService, err = GetUrlService(); err != nil {
			return nil, err
		}

		if requestService, err = GetRequestService(); err != nil {
			return nil, err
		}

		postServiceInstance = &PostService{
			fileService:    fileService,
			argService:     argService,
			jsonService:    jsonService,
			urlService:     urlService,
			requestService: requestService,
		}
	}

	return postServiceInstance, nil
}

func (ps *PostService) Run() {
	utils.Log.Info("Running GetService")

	if err := ps.GetUUID(); err != nil {
		utils.HandleErr(err)
	} else {
		utils.Log.Info("UUID generated", ps.UUID)
	}

	if err := ps.GetRequest(); err != nil {
		utils.HandleErr(err)
	} else {
		utils.Log.Info("Request parsed [", ps.Request["_ METHOD"], ps.Request["_ PATH"], "]")
	}

	if err := ps.GetPayloads(); err != nil {
		utils.HandleErr(err)
	} else {
		utils.Log.Info(fmt.Sprintf("Payloads parsed: [%d]", len(ps.Payloads)))
	}

	var combinedRequests []map[string]string
	for _, payload := range ps.Payloads {
		newRequest, err := ps.urlService.CombineRequestWithPayload(utils.CloneMap(ps.Request), payload)
		if err != nil {
			utils.HandleErr(err)
		}
		combinedRequests = append(combinedRequests, newRequest)
	}

	utils.Log.Info(fmt.Sprintf("Generated [%d] requests", len(combinedRequests)))
	ps.CombinedUrls = combinedRequests

	time.Sleep(2 * time.Second)
}

func (ps *PostService) GetRequest() error {
	var request map[string]string

	if requestFile, err := ps.argService.Get(ArgKeys.Urls); err != nil {
		return err
	} else {
		if fileContent, err := ps.fileService.ReadFileAsString(requestFile.(string)); err != nil {
			return err
		} else {
			if request, err = ps.urlService.ParseRequest(fileContent); err != nil {
				return err
			}
		}
	}

	ps.Request = request
	return nil
}

func (ps *PostService) GetPayloads() error {
	var payloads []string

	if payloadsFile, err := ps.argService.Get(ArgKeys.Payload); err != nil {
		return err
	} else {
		if fileContent, err := ps.fileService.ReadFileAsString(payloadsFile.(string)); err != nil {
			return err
		} else {
			payloads = strings.Split(fileContent, "\n")
		}
	}

	var validPayloads []string
	for _, payload := range payloads {
		if strings.Trim(payload, " ") == "" {
			continue
		}

		formattedPayload := strings.Replace(payload, "###", ps.UUID, -1)
		if _, err := url.ParseQuery(fmt.Sprintf("?id=%s", formattedPayload)); err != nil {
			validPayloads = append(validPayloads, url.QueryEscape(formattedPayload))
		} else {
			validPayloads = append(validPayloads, formattedPayload)
		}
	}

	ps.Payloads = validPayloads
	return nil
}

func (ps *PostService) GetUUID() error {
	dateNow := strconv.FormatInt(time.Now().UnixMicro(), 10)
	ps.UUID = strings.Replace(dateNow, ".", "", -1)

	return nil
}

func (ps *PostService) Scan(ctx playwright.BrowserContext) error {
	options := ps.argService.GetAll()

	ctx.NewPage()

	continueFrom := min(max(0, options[ArgKeys.Continue].(int)), len(ps.CombinedUrls))
	urlsToScan := ps.CombinedUrls[continueFrom:]

	var wg sync.WaitGroup

	tabCount := min(options[ArgKeys.Threads].(int), len(urlsToScan))

	// for ix := 0; ix < tabCount; ix++ {
	// 	if _, err := ctx.NewPage(); err != nil {
	// 		utils.HandleErr(err)
	// 	}
	// }

	urlsChan := make(chan map[string]string, tabCount)

	if continueFrom > 0 {
		utils.Log.Info(fmt.Sprintf("Continuing from index: %d", continueFrom))
	}

	doneCtx, cancel := context.WithCancel(context.Background())

	go func() {
	outer:
		for ix, url := range urlsToScan {
			wg.Add(1)
			urlsChan <- url

			select {
			case <-doneCtx.Done():
				utils.Log.Warn("Ending the scan, last url index: ", ix)
				break outer
			default:
				if ix%options[ArgKeys.Threads].(int) == 0 {
					utils.Log.Info(fmt.Sprintf("Scanning URL progress: %d/%d %d%%", ix+1, len(urlsToScan), (ix+1)*100/len(urlsToScan)))
				}
			}
		}
		close(urlsChan)
	}()

	freeSlotChan := make(chan int, tabCount)
	for i := 0; i < tabCount; i++ {
		freeSlotChan <- i
	}

	var foundXss []map[string]string
	var m sync.Mutex

	utils.HandleCtrlC(cancel)

	for {
		url, ok := <-urlsChan
		if !ok {
			break
		}

		shouldBreak := false
		select {
		case <-doneCtx.Done():
			wg.Done()
			shouldBreak = true
		default:
		}

		if shouldBreak {
			utils.Log.Warn("Exiting the application, please wait few seconds for cleanup")
			// exhaust urls channels
			for range urlsChan {
				wg.Done()
			}
			break
		}

		pageIndex, ok := <-freeSlotChan
		if pageIndex == -1 || !ok || shouldBreak {
			break
		}
		go func(url map[string]string, pageIndex int) {
			if found, err := ps.Send(ctx, url, pageIndex); err != nil {
				utils.Log.Error(fmt.Sprintf("Error sending request: %s", err))
			} else if found {
				m.Lock()
				foundXss = append(foundXss, url)
				utils.Log.Success(fmt.Sprintf("Found total %d urls", len(foundXss)))
				m.Unlock()
			}

			if options[ArgKeys.Delay].(int) > 0 {
				time.Sleep(time.Duration(options[ArgKeys.Delay].(int)) * time.Millisecond)
			}

			freeSlotChan <- pageIndex
			wg.Done()

		}(url, pageIndex)
	}

	utils.Log.Error("Waiting for all requests to finish")
	wg.Wait()
	utils.Log.Error("All requests finished")

	close(freeSlotChan)

	defer func() {
		if len(foundXss) > 0 {
			utils.Log.Success(fmt.Sprintf("XSS found: %d", len(foundXss)))
			for _, url := range foundXss {
				utils.Log.Success(url["_ BODY"])
			}
			utils.Log.Success(fmt.Sprintf("XSS found: %d", len(foundXss)))
		} else {
			utils.Log.Info("No XSS found")
		}

		if options[ArgKeys.Report].(string) != "" {
			foundXssBodies := func(s []map[string]string) []string {
				var foundXss []string
				for _, url := range s {
					foundXss = append(foundXss, url["_ BODY"])
				}
				return foundXss
			}(foundXss)

			if foundXssJson, err := ps.jsonService.ArrayToString(foundXssBodies); err != nil {
				utils.HandleErr(err)
			} else {
				if err := ps.fileService.WriteFileAsString(options[ArgKeys.Report].(string), foundXssJson); err != nil {
					utils.HandleErr(err)
				} else {
					utils.Log.Info(fmt.Sprintf("Report saved to: %s", options[ArgKeys.Report].(string)))
				}
			}
		}
	}()

	defer func() {
		if err := recover(); err != nil {
			utils.HandlePanic(false)
		}
	}()

	return nil
}

func (ps *PostService) Send(ctx playwright.BrowserContext, url map[string]string, pageIndex int) (foundXss bool, _ error) {
	options := ps.argService.GetAll()

	var page playwright.Page
	var err error

	page, err = ctx.NewPage()
	if err != nil {
		return false, err
	}

	fullUrl := fmt.Sprintf("%s://%s%s", options[ArgKeys.Protocol], url["Host"], url["_ PATH"])

	dialogChan := make(chan bool, 1)
	pageLoadChan := make(chan bool, 1)

	handleDialog := func(dialog playwright.Dialog) {
		dialogMsg := dialog.Message()
		dialogType := dialog.Type()
		if err := dialog.Accept(); err != nil {
			utils.Log.Error(fmt.Sprintf("Error accepting dialog %s: %s, %s", dialogType, dialogMsg, url["_ BODY"]))
		} else if !foundXss {
			if dialogMsg == ps.UUID {
				utils.Log.Success(fmt.Sprintf("XSS found: %s", url["_ BODY"]))
			} else {
				utils.Log.Warn(fmt.Sprintf("Alert found with UUID missmatch: %s %s", dialogMsg, url["_ BODY"]))
			}
			foundXss = true
			dialogChan <- true
		}
		// else if page != nil {
		// 	_, err := page.Evaluate(`() => {
		// 		window.alert = () => {};
		// 		window.confirm = () => true;
		// 		window.prompt = () => null;
		// 	}`)
		// 	utils.HandleErr(err)
		// }
	}

	handlePageLoad := func(page playwright.Page) {
		pageLoadChan <- true
	}

	page.On("dialog", handleDialog)
	page.On("load", handlePageLoad)

	err = page.Route(fullUrl, func(route playwright.Route) {
		// send post request
		response, respHeaders, err := ps.requestService.Post(fullUrl, url["_ BODY"], url)
		if err != nil {
			utils.Log.Error(fmt.Sprintf("Error sending request: %s", err))
			return
		}

		route.Fulfill(playwright.RouteFulfillOptions{
			Body:    response,
			Headers: respHeaders,
		})
	})

	if err != nil {
		return false, err
	}

	defer func() {
		if page != nil {
			page.Close()
			page.RemoveListener("dialog", handleDialog)
			page.RemoveListener("load", handlePageLoad)
		}
		if dialogChan != nil {
			close(dialogChan)
		}
		if pageLoadChan != nil {
			close(pageLoadChan)
		}
	}()

	if _, err := page.Goto(fullUrl); err != nil {
		return false, err
	}

	select {
	case <-pageLoadChan:
		// Page loaded
	case <-time.After(time.Duration(options[ArgKeys.Timeout].(int)) * time.Millisecond):
		utils.Log.Warn(fmt.Sprintf("Timeout loading page: %s", url))
		return false, nil
	}

	select {
	case <-dialogChan:
		page.Close()
		page = nil
		return true, nil
	case <-time.After(time.Duration(1000) * time.Millisecond):
	}

	return false, nil
}
