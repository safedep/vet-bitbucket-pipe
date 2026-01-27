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
				foundThreats := false
				foundViolations := false

				for _, data := range report.Data {
					switch data.Title {
					case "Safe to Merge":
						foundSafeToMerge = true
						assert.Equal(t, false, data.Value, "Safe to Merge value mismatch")
					case "Malicious Packages":
						foundMalicious = true
						assert.Equal(t, 1, data.Value, "Malicious Packages count mismatch")
					case "Suspicious Packages":
						foundSuspicious = true
						assert.Equal(t, 4, data.Value, "Suspicious Packages count mismatch")
					case "Vulnerabilities":
						foundVulnerabilities = true
						assert.Equal(t, 10, data.Value, "Vulnerabilities count mismatch")
					case "Threats":
						foundThreats = true
						assert.Equal(t, 0, data.Value, "Threats count mismatch")
					case "Violations":
						foundViolations = true
						assert.Equal(t, 1, data.Value, "Violations count mismatch")
					}
				}

				assert.True(t, foundSafeToMerge, "expected to find 'Safe to Merge' data point")
				assert.True(t, foundMalicious, "expected to find 'Malicious Packages' data point")
				assert.True(t, foundSuspicious, "expected to find 'Suspicious Packages' data point")
				assert.True(t, foundVulnerabilities, "expected to find 'Vulnerabilities' data point")
				assert.True(t, foundThreats, "expected to find 'Threats' data point")
				assert.True(t, foundViolations, "expected to find 'Violations' data point")
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

	vulnCount := 0
	maliciousCount := 0
	suspiciousCount := 0
	violationCount := 0
	threatCount := 0

	testAnnotationsLinks := []string{}

	for _, annotation := range annotations {
		switch annotation.AnnotationType {
		case AnnotationTypeVulnerability:
			vulnCount++
		case AnnotationTypeCodeSmell:
			violationCount++
		case AnnotationTypeBug:
			if annotation.Title == "Malicious Package" {
				maliciousCount++
			} else if annotation.Title == "Suspicious Package" {
				suspiciousCount++
			} else if annotation.Title == "Threat Detected" {
				threatCount++
			}
		}

		testAnnotationsLinks = append(testAnnotationsLinks, annotation.Link)
	}

	assert.Equal(t, 10, vulnCount, "Expected 10 vulnerability annotations")
	assert.Equal(t, 1, maliciousCount, "Expected 1 malicious package annotation")
	assert.Equal(t, 4, suspiciousCount, "Expected 4 suspicious package annotations")
	assert.Equal(t, 1, violationCount, "Expected 1 violation annotation")
	assert.Equal(t, 0, threatCount, "Expected 0 threat annotation")

	assert.Contains(t, testAnnotationsLinks, "https://app.safedep.io/community/malysis/01KFDHJ20PFG657733QSZYBGM9")
	assert.Contains(t, testAnnotationsLinks, "https://app.safedep.io/community/malysis/01JQ608SG8GMYG28E9SA0BPBTR")
	assert.Contains(t, testAnnotationsLinks, "https://app.safedep.io/community/malysis/01JQ5XJ30Z5A4Z82H6TVVFH400")
	assert.Contains(t, testAnnotationsLinks, "https://app.safedep.io/community/malysis/01JQ5Y0C8NP98F0BEZQC5RW2PP")
}
