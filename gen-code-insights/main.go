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

	if err != nil {
		fmt.Printf("failed to initiate bitbucket code insights generator: %v\n", err)
		os.Exit(1)
	}

	ciReport, err := ciGenerator.GenerateReport()
	if err != nil {
		fmt.Printf("failed to generate code insights report: %v\n", err)
		os.Exit(1)
	}

	ciAnnotations, err := ciGenerator.GenerateAnnotations()
	if err != nil {
		fmt.Printf("failed to generate code insights annotations: %v\n", err)
		os.Exit(1)
	}

	if err := SaveModel(ciReport, codeInsightsReportJsonFilePath); err != nil {
		fmt.Printf("failed to save code insights report data: %v\n", err)
		os.Exit(1)
	}

	if err := SaveModel(ciAnnotations, codeInsightsAnnotationsJsonFilePath); err != nil {
		fmt.Printf("failed to save code insights report data: %v\n", err)
		os.Exit(1)
	}
}
