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

	playwright "github.com/playwright-community/playwright-go"
)

type GetService struct {
	CombinedUrls []string
	Urls         []string
	Payloads     []string
	UUID         string

	fileService *FileService
	argService  *ArgsService
	jsonService *JsonService
	urlService  *UrlService
}

var getServiceInstance *GetService = nil

// Singleton instance of GetService
func GetGetService() (*GetService, error) {
	if getServiceInstance == nil {
		var fileService *FileService
		var argService *ArgsService
		var jsonService *JsonService
		var urlService *UrlService
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

		getServiceInstance = &GetService{
			fileService: fileService,
			argService:  argService,
			jsonService: jsonService,
			urlService:  urlService,
		}
	}

	return getServiceInstance, nil
}

func (gs *GetService) Run() {
	utils.Log.Info("Running GetService")

	if err := gs.GetUUID(); err != nil {
		utils.HandleErr(err)
	} else {
		utils.Log.Info("UUID generated", gs.UUID)
	}

	if err := gs.GetUrls(); err != nil {
		utils.HandleErr(err)
	} else {
		utils.Log.Info(fmt.Sprintf("URLs parsed: [%d]", len(gs.Urls)))
	}

	if err := gs.GetPayloads(); err != nil {
		utils.HandleErr(err)
	} else {
		utils.Log.Info(fmt.Sprintf("Payloads parsed: [%d]", len(gs.Payloads)))
	}

	var combinedUrls []string
	for _, url := range gs.Urls {
		newUrls := gs.urlService.CombineUrlQueryWithPayload(url, gs.Payloads)
		combinedUrls = append(combinedUrls, newUrls...)
	}

	utils.Log.Info(fmt.Sprintf("Generated [%d] URLs", len(combinedUrls)))
	gs.CombinedUrls = utils.RemoveDuplicates(combinedUrls)

	time.Sleep(2 * time.Second)
}

func (gs *GetService) GetUrls() error {
	var urls []string

	if urlsFile, err := gs.argService.Get(ArgKeys.Urls); err != nil {
		return err
	} else {
		if fileContent, err := gs.fileService.ReadFileAsString(urlsFile.(string)); err != nil {
			return err
		} else {
			urls = strings.Split(fileContent, "\n")
		}
	}

	var validUrls []string
	for _, url := range urls {
		if strings.Trim(url, " ") == "" {
			continue
		}

		if gs.urlService.ValidateUrl(url) {
			validUrls = append(validUrls, url)
		}
	}

	gs.Urls = validUrls
	return nil
}

func (gs *GetService) GetPayloads() error {
	var payloads []string

	if payloadsFile, err := gs.argService.Get(ArgKeys.Payload); err != nil {
		return err
	} else {
		if fileContent, err := gs.fileService.ReadFileAsString(payloadsFile.(string)); err != nil {
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

		formattedPayload := strings.Replace(payload, "###", gs.UUID, -1)
		if _, err := url.ParseQuery(fmt.Sprintf("?id=%s", formattedPayload)); err != nil {
			validPayloads = append(validPayloads, url.QueryEscape(formattedPayload))
		} else {
			validPayloads = append(validPayloads, formattedPayload)
		}
	}

	gs.Payloads = validPayloads
	return nil
}

func (gs *GetService) GetUUID() error {
	dateNow := strconv.FormatInt(time.Now().UnixMicro(), 10)
	gs.UUID = strings.Replace(dateNow, ".", "", -1)

	return nil
}

func (gs *GetService) Send(ctx playwright.BrowserContext, url string, pageIndex int) (foundXss bool, _ error) {
	options := gs.argService.GetAll()

	var page playwright.Page
	var err error

	page, err = ctx.NewPage()
	if err != nil {
		return false, err
	}

	dialogChan := make(chan bool, 1)
	pageLoadChan := make(chan bool, 1)

	handleDialog := func(dialog playwright.Dialog) {
		dialogMsg := dialog.Message()
		dialogType := dialog.Type()
		if err := dialog.Accept(); err != nil {
			utils.Log.Error(fmt.Sprintf("Error accepting dialog %s: %s, %s", dialogType, dialogMsg, url))
		} else {
			if dialogMsg == gs.UUID {
				utils.Log.Success(fmt.Sprintf("XSS found: %s", url))
			} else {
				utils.Log.Warn(fmt.Sprintf("Alert found with UUID missmatch: %s %s", dialogMsg, url))
			}
			foundXss = true
			dialogChan <- true
		}
	}

	handlePageLoad := func(page playwright.Page) {
		pageLoadChan <- true
	}

	page.On("dialog", handleDialog)
	page.On("load", handlePageLoad)

	defer func() {
		if page != nil {
			page.RemoveListener("dialog", handleDialog)
			page.RemoveListener("load", handlePageLoad)
			if dialogChan != nil {
				close(dialogChan)
			}
			if pageLoadChan != nil {
				close(pageLoadChan)
			}
		}
		page.Close()
	}()

	if _, err := page.Goto(url); err != nil {
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
		return true, nil
	case <-time.After(time.Duration(1000) * time.Millisecond):
		//utils.Log.Warn(fmt.Sprintf("1000ms timeout waiting for dialog: %s", url))
	}

	return false, nil
}

func (gs *GetService) Scan(ctx playwright.BrowserContext) error {
	options := gs.argService.GetAll()

	ctx.NewPage()

	continueFrom := min(max(0, options[ArgKeys.Continue].(int)), len(gs.CombinedUrls))
	urlsToScan := gs.CombinedUrls[continueFrom:]

	var wg sync.WaitGroup

	tabCount := min(options[ArgKeys.Threads].(int), len(urlsToScan))

	urlsChan := make(chan string, tabCount)

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

	var foundXss []string
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
		go func(url string, pageIndex int) {
			if found, err := gs.Send(ctx, url, pageIndex); err != nil {
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
				utils.Log.Success(url)
			}
			utils.Log.Success(fmt.Sprintf("XSS found: %d", len(foundXss)))
		} else {
			utils.Log.Info("No XSS found")
		}

		if options[ArgKeys.Report].(string) != "" {
			if foundXssJson, err := gs.jsonService.ArrayToString(foundXss); err != nil {
				utils.HandleErr(err)
			} else {
				if err := gs.fileService.WriteFileAsString(options[ArgKeys.Report].(string), foundXssJson); err != nil {
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
