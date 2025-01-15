package app

import (
	"strings"
	"xss/services"
	"xss/utils"

	"github.com/playwright-community/playwright-go"
)

type App struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	ctx     playwright.BrowserContext

	argService  *services.ArgsService
	getService  *services.GetService
	postService *services.PostService
}

func NewApp() *App {
	var argService *services.ArgsService
	var getService *services.GetService
	var postService *services.PostService
	var err error

	if argService, err = services.GetArgsService(); err != nil {
		utils.HandleErr(err)
	}

	if getService, err = services.GetGetService(); err != nil {
		utils.HandleErr(err)
	}

	if postService, err = services.GetPostService(); err != nil {
		utils.HandleErr(err)
	}

	return &App{
		argService:  argService,
		getService:  getService,
		postService: postService,
	}
}

func (app *App) Start() error {
	utils.Log.Info("Starting the app")
	if instance, err := playwright.Run(&playwright.RunOptions{}); err != nil {
		return err
	} else {
		app.pw = instance
		utils.Log.Info("Instance started")
	}

	utils.Log.Info("Launching the browser")
	if browser, err := app.pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless:     playwright.Bool(false),
		HandleSIGINT: playwright.Bool(false),
	}); err != nil {
		return err
	} else {
		app.browser = browser
		utils.Log.Info("Browser launched")
	}

	if ctx, err := app.browser.NewContext(); err != nil {
		return err
	} else {
		app.ctx = ctx
	}

	return nil
}

func (app *App) Close() {
	utils.Log.Info("Closing the app")
	if err := app.ctx.Close(); err != nil {
		utils.HandleErr(err)
	} else {
		utils.Log.Info("Context closed")
	}

	utils.Log.Info("Closing the browser")
	if err := app.browser.Close(); err != nil {
		utils.HandleErr(err)
	}
	utils.Log.Info("Browser closed")

	utils.Log.Info("Stopping the instance")
	if err := app.pw.Stop(); err != nil {
		utils.HandleErr(err)
	}
	utils.Log.Info("Instance stopped")
}

func (app *App) Run() {
	utils.Log.Success("Running the app")

	method, err := app.argService.Get(services.ArgKeys.Method)
	utils.HandleErr(err)

	if strings.ToUpper(method.(string)) == "GET" {
		app.getService.Run()
		err := app.getService.Scan(app.ctx)
		utils.HandleErr(err)
	} else {
		app.postService.Run()
	}

	utils.Log.Success("Ran the app")
}
