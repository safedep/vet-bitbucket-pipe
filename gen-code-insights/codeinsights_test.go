package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCodeInsightsGenerator(t *testing.T) {
	wf, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
		ReportTitle:          "Test Dependency Scanning",
		ReportVendor:         "safedep/vet-bitbucket-pipe",
		SourceJsonReportFile: "testdata/test-json-report.json",
	})

	assert.NoError(t, err)
	assert.NotNil(t, wf.reportData)
}

func TestGenerateReport(t *testing.T) {
	reportTitle := "Test Dependency Scanning"
	reportVendor := "safedep/vet-bitbucket-pipe"

	wf, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
		ReportTitle:          reportTitle,
		ReportVendor:         reportVendor,
		SourceJsonReportFile: "testdata/test-json-report.json",
	})

	assert.NoError(t, err)

	report, err := wf.GenerateReport()
	assert.NoError(t, err)

	assert.Equal(t, reportTitle, report.Title)
	assert.Equal(t, reportVendor, report.Reporter)
	assert.Equal(t, ReportResultFailed, report.Result)

	foundSafeToMerge := false
	foundMalicious := false
	foundSuspicious := false
	foundVulnerabilities := false

	for _, data := range report.Data {
		switch data.Title {
		case "Safe to Merge":
			foundSafeToMerge = true
			assert.Equal(t, report.Result, data.Value, "Safe to Merge value mismatch")
		case "Malicious Packages":
			foundMalicious = true
			assert.Equal(t, 1, data.Value, "Malicious Packages count mismatch")
		case "Suspicious Packages":
			foundSuspicious = true
			assert.Equal(t, 0, data.Value, "Suspicious Packages count mismatch")
		case "Vulnerabilities":
			foundVulnerabilities = true
			assert.Equal(t, 10, data.Value, "Vulnerabilities count mismatch")
		}
	}

	assert.True(t, foundSafeToMerge, "expected to find 'Safe to Merge' data point")
	assert.True(t, foundMalicious, "expected to find 'Malicious Packages' data point")
	assert.True(t, foundSuspicious, "expected to find 'Suspicious Packages' data point")
	assert.True(t, foundVulnerabilities, "expected to find 'Vulnerabilities' data point")
}

func TestGenerateAnnotations(t *testing.T) {
	wf, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
		SourceJsonReportFile: "testdata/test-json-report.json",
	})

	assert.NoError(t, err)

	annotations, err := wf.GenerateAnnotations()
	assert.NoError(t, err)

	expectedAnnotations := 11 // 10 vulnerabilities + 1 malicious package
	assert.Equal(t, expectedAnnotations, len(*annotations))
}
