package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/safedep/dry/api/pb"
	jsonreportspec "github.com/safedep/vet/gen/jsonreport"
	"github.com/safedep/vet/gen/models"
)

type codeInsightsGenerator struct {
	reportData *jsonreportspec.Report
	config     CodeInsightsGeneratorConfig
}

type CodeInsightsGeneratorConfig struct {
	ReportTitle          string
	ReportVendor         string
	SourceJsonReportFile string
}

func NewCodeInsightsGenerator(config CodeInsightsGeneratorConfig) (codeInsightsGenerator, error) {
	if config.SourceJsonReportFile == "" {
		return codeInsightsGenerator{}, fmt.Errorf("source json report file is missing")
	}

	data, err := os.ReadFile(config.SourceJsonReportFile)
	if err != nil {
		return codeInsightsGenerator{}, fmt.Errorf("failed to read source json report file: %w", err)
	}

	report := new(jsonreportspec.Report)
	err = pb.FromJson(bytes.NewReader(data), report)

	return codeInsightsGenerator{
		config:     config,
		reportData: report,
	}, nil
}

// GenerateReport is only concerned about Summary of the report
func (ci codeInsightsGenerator) GenerateReport() (*CodeInsightsReport, error) {
	vulnerabilitiesCnt := 0
	violationsCnt := 0
	threatsCnt := 0
	maliciousCnt := 0
	suspiciousCnt := 0

	for _, pkg := range ci.reportData.GetPackages() {
		vulnerabilitiesCnt += len(pkg.GetVulnerabilities())
		violationsCnt += len(pkg.GetViolations())
		threatsCnt += len(pkg.GetThreats())

		for _, m := range pkg.GetMalwareInfo() {
			switch m.GetType() {
			case jsonreportspec.MalwareType_SUSPICIOUS:
				suspiciousCnt++
			case jsonreportspec.MalwareType_MALICIOUS:
				maliciousCnt++
			}
		}
	}

	report := &CodeInsightsReport{
		Title:      ci.config.ReportTitle,
		Reporter:   ci.config.ReportVendor,
		ReportType: ReportTypeSecurity,
		Result:     ReportResultPassed,
		Data:       make([]CodeInsightsData, 0),
	}

	// OPNIONATED
	// If anything exists, vulnerability or suspicious or malicious packages
	// then the PR is not "Safe" to Merge
	if (vulnerabilitiesCnt +
		violationsCnt +
		threatsCnt +
		suspiciousCnt +
		maliciousCnt) > 0 {
		report.Result = ReportResultFailed
	}

	report.Details = ""

	if report.Result == ReportResultFailed {
		report.Details = ""
	}

	// Safe to merge data point
	report.Data = append(report.Data, CodeInsightsData{
		Title: "Safe to Merge",
		Type:  DataTypeBoolean,
		Value: report.Result,
	})

	// Key Value Data Points to show in report UI
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Malicious Packages", maliciousCnt))
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Suspicious Packages", threatsCnt))
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Vulnerabilities", vulnerabilitiesCnt))
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Threats", threatsCnt))
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Violations", violationsCnt))

	return report, nil
}

func (ci codeInsightsGenerator) GenerateAnnotations() (*[]CodeInsightsAnnotation, error) {
	annotations := []CodeInsightsAnnotation{}

	for _, pkg := range ci.reportData.GetPackages() {
		manifestPath := ""
		if len(pkg.GetManifests()) > 0 {
			for _, m := range ci.reportData.GetManifests() {
				if m.GetId() == pkg.GetManifests()[0] {
					manifestPath = m.GetPath()
					break
				}
			}
		}

		for _, vuln := range pkg.GetVulnerabilities() {
			severity := AnnotationSeverityLow
			if len(vuln.GetSeverities()) > 0 {
				sev := vuln.GetSeverities()[0]
				switch sev.GetRisk() {
				case models.InsightVulnerabilitySeverity_CRITICAL:
					severity = AnnotationSeverityCritical
				case models.InsightVulnerabilitySeverity_HIGH:
					severity = AnnotationSeverityHigh
				case models.InsightVulnerabilitySeverity_MEDIUM:
					severity = AnnotationSeverityMedium
				case models.InsightVulnerabilitySeverity_LOW:
					severity = AnnotationSeverityLow
				}
			}

			annotations = append(annotations, CodeInsightsAnnotation{
				ExternalID:     vuln.GetId(),
				Title:          vuln.GetTitle(),
				AnnotationType: AnnotationTypeVulnerability,
				Summary:        vuln.GetTitle(),
				Severity:       severity,
				FilePath:       manifestPath,
			})
		}

		for _, m := range pkg.GetMalwareInfo() {
			if m.GetType() == jsonreportspec.MalwareType_MALICIOUS {
				annotations = append(annotations, CodeInsightsAnnotation{
					Title:          "Malicious Package",
					AnnotationType: AnnotationTypeBug,
					Summary:        fmt.Sprintf("Package %s@%s is malicious", pkg.GetPackage().GetName(), pkg.GetPackage().GetVersion()),
					Severity:       AnnotationSeverityCritical,
					FilePath:       manifestPath,
				})
			}
		}
	}

	return &annotations, nil
}

func createNumericCodeInsightsDataPoint(title string, value int) CodeInsightsData {
	return CodeInsightsData{
		Title: title,
		Type:  DataTypeNumber,
		Value: value,
	}
}
