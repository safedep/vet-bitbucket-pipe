package main

import (
	"flag"
	"fmt"
	"os"
)

const (
	// These file names are contract at `pipe/upload_report.sh`
	codeInsightsReportJsonFilePath      = "code-insights-report.json"
	codeInsightsAnnotationsJsonFilePath = "code-insights-annotations.json"
)

func main() {
	var jsonReportFile string

	flag.StringVar(&jsonReportFile, "json-report-file", "", "Vet generated JSON report file path")
	flag.Parse()

	// ci == Code Insights
	ciGenerator, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
		SourceJsonReportFile: jsonReportFile,
		ReportTitle:          "SafeDep Dependency Scan",
		ReportVendor:         "safedep.io",
	})
	handleError("failed to initiate bitbucket code insights generator", err)

	ciReport, err := ciGenerator.GenerateReport()
	handleError("failed to generate code insights report", err)

	ciAnnotations, err := ciGenerator.GenerateAnnotations()
	handleError("failed to generate code insights annotations", err)

	err = SaveModel(ciReport, codeInsightsReportJsonFilePath)
	handleError("failed to save code insights report data", err)

	err = SaveModel(ciAnnotations, codeInsightsAnnotationsJsonFilePath)
	handleError("failed to save code insights report data", err)
}

func handleError(msg string, err error) {
	if err != nil {
		fmt.Printf("%s: %v", msg, err)
		os.Exit(1)
	}
}
