package main

import (
	"time"
	"xss/app"
	"xss/services"
	"xss/utils"
)

func main() {
	start := time.Now()

	fileService, err := services.GetFileService()
	utils.HandleErr(err)

	args, err := services.GetArgsService()
	utils.HandleErr(err)

	options := args.GetAll()
	utils.HandleErr(err)

	defer func() {
		if len(options[services.ArgKeys.Command].(string)) == 0 {
			utils.Log.Warn("No command provided, use 'scan' or 'report'")
			utils.Log.Warn("See -h/--help for more information")
		} else {
			utils.Log.Custom(utils.Colors.Cyan, "TIMER", "Execution time:", time.Since(start))
			utils.Log.Info("Exiting the application")
		}
	}()

	if val, err := args.Get(services.ArgKeys.Command); err != nil {
		utils.HandleErr(err)
	} else if val.(string) == "scan" {
		app := app.NewApp()
		app.Start()
		app.Run()
		defer app.Close()
	} else if val.(string) == "report" {
		reportService, err := services.GetReportService()
		utils.HandleErr(err)

		err = reportService.DisplayReport(options[services.ArgKeys.Report].(string))
		utils.HandleErr(err)
	}

	defer func() {
		output, err := args.Get(services.ArgKeys.Output)

		if err != nil {
			return
		}

		if output != "" {
			utils.Log.Info("Writing logs to file ", output)
			msgs := utils.Log.GetMessages()
			err := fileService.WriteFileAsJSON(output.(string), msgs)
			utils.HandleErr(err)
		}
	}()
	defer utils.HandlePanic(false)
}
