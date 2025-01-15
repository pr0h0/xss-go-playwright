package services

import "xss/utils"

type ReportService struct {
	jsonService *JsonService
	fileService *FileService
}

var reportServiceInstance *ReportService

// GetReportService returns the singleton instance of the ReportService
func GetReportService() (*ReportService, error) {
	if reportServiceInstance == nil {
		jsonService, err := GetJsonService()
		if err != nil {
			utils.HandleErr(err)
		}

		fileService, err := GetFileService()
		if err != nil {
			utils.HandleErr(err)
		}

		reportServiceInstance = &ReportService{
			jsonService: jsonService,
			fileService: fileService,
		}
	}
	return reportServiceInstance, nil
}

// DisplayReport reads the report file and displays the report
func (rs *ReportService) DisplayReport(reportFile string) error {
	var report []string

	err := rs.fileService.ReadFileAsJSON(reportFile, &report)
	if err != nil {
		return err
	}
	utils.Log.Info("Displaying report from", reportFile)
	for _, msg := range report {
		utils.Log.Success(msg)
	}

	return nil
}
