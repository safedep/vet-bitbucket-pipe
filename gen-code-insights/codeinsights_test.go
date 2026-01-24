package main

import (
	"testing"
)

func TestNewCodeInsightsGenerator(t *testing.T) {
	wf, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
		ReportTitle:          "Test Dependency Scanning",
		ReportVendor:         "safedep/vet-bitbucket-pipe",
		SourceJsonReportFile: "testdata/test-json-report.json",
	})

	if err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	if wf.reportData == nil {
		t.Fatal("report data is nil")
	}
}

func TestGenerateReport(t *testing.T) {
	wf, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
		SourceJsonReportFile: "testdata/test-json-report.json",
	})

	if err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	report, err := wf.GenerateReport()
	if err != nil {
		t.Fatalf("failed to generate report: %v", err)
	}

	if report.Result != ReportResultFailed {
		t.Errorf("expected report result to be FAILED, got %s", report.Result)
	}
}

func TestGenerateAnnotations(t *testing.T) {
	wf, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
		SourceJsonReportFile: "testdata/test-json-report.json",
	})

	if err != nil {
		t.Fatalf("failed to create workflow: %v", err)
	}

	annotations, err := wf.GenerateAnnotations()
	if err != nil {
		t.Fatalf("failed to generate annotations: %v", err)
	}

	if len(*annotations) == 0 {
		t.Fatal("no annotations generated")
	}
}
