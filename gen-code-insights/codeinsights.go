package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/safedep/dry/api/pb"
	jsonreportspec "github.com/safedep/vet/gen/jsonreport"
)

type CodeInsightsGeneratorWorkflow struct {
	ReportData *jsonreportspec.Report
}

type CodeInsightsGeneratorWorkflowConfig struct {
	SourceJsonReportFile string
}

func NewCodeInsightsGeneratorWorkflow(config CodeInsightsGeneratorWorkflowConfig) (CodeInsightsGeneratorWorkflow, error) {
	if config.SourceJsonReportFile == "" {
		return CodeInsightsGeneratorWorkflow{}, fmt.Errorf("source json report file is missing")
	}

	data, err := os.ReadFile(config.SourceJsonReportFile)
	if err != nil {
		return CodeInsightsGeneratorWorkflow{}, fmt.Errorf("failed to read source json report file: %w", err)
	}

	report := new(jsonreportspec.Report)
	err = pb.FromJson(bytes.NewReader(data), report)

	return CodeInsightsGeneratorWorkflow{
		ReportData: report,
	}, nil
}

func (ci CodeInsightsGeneratorWorkflow) GenerateReport() (*CodeInsightsReport, error) {
	return nil, nil
}

func (ci CodeInsightsGeneratorWorkflow) GenerateAnnotations() (*CodeInsightsAnnotation, error) {
	return nil, nil
}
