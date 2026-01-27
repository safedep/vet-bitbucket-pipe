package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var jsonReportFile string
	var codeInsightsReportJsonFilePath string
	var codeInsightsAnnotationsJsonFilePath string

	flag.StringVar(&jsonReportFile, "source-json-report-file", "", "Vet generated JSON report file path")
	flag.StringVar(&codeInsightsReportJsonFilePath, "dest-report-file", "code-insights-report.json", "Bitbucket Code Insights report file path")
	flag.StringVar(&codeInsightsAnnotationsJsonFilePath, "dest-annotations-file", "code-insights-annotations.json", "Bitbucket Code Insights annotations file path")

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
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}
