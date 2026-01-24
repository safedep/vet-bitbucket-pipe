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

	testCases := []struct {
		name                 string
		sourceJsonReportFile string
		expectedResult       ReportResult
		expectedDetails      string
	}{
		{
			name:                 "FAILED",
			sourceJsonReportFile: "testdata/test-json-report.json",
			expectedResult:       ReportResultFailed,
			expectedDetails:      "Issues found, please check the report for details.",
		},
		{
			name:                 "PASSED",
			sourceJsonReportFile: "testdata/test-json-report-pass.json",
			expectedResult:       ReportResultPassed,
			expectedDetails:      "No issues found.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			wf, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
				ReportTitle:          reportTitle,
				ReportVendor:         reportVendor,
				SourceJsonReportFile: tc.sourceJsonReportFile,
			})
			assert.NoError(t, err)

			report, err := wf.GenerateReport()
			assert.NoError(t, err)

			assert.Equal(t, reportTitle, report.Title)
			assert.Equal(t, reportVendor, report.Reporter)
			assert.Equal(t, tc.expectedResult, report.Result)
			assert.Equal(t, tc.expectedDetails, report.Details)

			if tc.name == "FAILED" {
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
		})
	}
}

func TestGenerateAnnotations(t *testing.T) {
	wf, err := NewCodeInsightsGenerator(CodeInsightsGeneratorConfig{
		SourceJsonReportFile: "testdata/test-json-report.json",
	})

	assert.NoError(t, err)

	annotations, err := wf.GenerateAnnotations()
	assert.NoError(t, err)

	expectedAnnotations := 21 // 10 vulnerabilities + 1 malicious package + 4 suspicious packages + 1 violation + 5 threats
	assert.Equal(t, expectedAnnotations, len(*annotations))
}