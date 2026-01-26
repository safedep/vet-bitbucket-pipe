package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"

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
	if err := pb.FromJson(bytes.NewReader(data), report); err != nil {
		return codeInsightsGenerator{}, fmt.Errorf("failed to read protobuf json spec: %w", err)
	}

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

	// OPINIONATED
	// If anything exists, vulnerability or malicious packages
	// then the PR is not "Safe" to Merge
	if (vulnerabilitiesCnt +
		violationsCnt +
		threatsCnt +
		maliciousCnt) > 0 {
		report.Result = ReportResultFailed
	}

	if report.Result == ReportResultFailed {
		report.Details = "Issues found, please check the report for details."
	} else {
		if suspiciousCnt > 0 {
			report.Details = fmt.Sprintf("Found %d suspicious packages, human review is recommended", suspiciousCnt)
		} else {
			report.Details = "No issues found."
		}
	}

	// Safe to merge data point
	report.Data = append(report.Data, CodeInsightsData{
		Title: "Safe to Merge",
		Type:  DataTypeBoolean,
		Value: report.Result == ReportResultPassed,
	})

	// Key Value Data Points to show in report UI
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Malicious Packages", maliciousCnt))
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Suspicious Packages", suspiciousCnt))
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Vulnerabilities", vulnerabilitiesCnt))
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Threats", threatsCnt))
	report.Data = append(report.Data, createNumericCodeInsightsDataPoint("Violations", violationsCnt))

	return report, nil
}

func (ci codeInsightsGenerator) GenerateAnnotations() ([]CodeInsightsAnnotation, error) {
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

		for _, m := range pkg.GetMalwareInfo() {
			threatLink := "https://app.safedep.io/community/malysis/" + strings.TrimPrefix(m.GetThreatId(), "SD-MAL-")

			switch m.GetType() {
			case jsonreportspec.MalwareType_MALICIOUS:
				annotations = append(annotations, CodeInsightsAnnotation{
					Title:          "Malicious Package",
					AnnotationType: AnnotationTypeBug,
					Summary:        fmt.Sprintf("Package %s@%s is malicious", pkg.GetPackage().GetName(), pkg.GetPackage().GetVersion()),
					Severity:       AnnotationSeverityCritical,
					FilePath:       manifestPath,
					Link:           threatLink,
				})
			case jsonreportspec.MalwareType_SUSPICIOUS:
				annotations = append(annotations, CodeInsightsAnnotation{
					Title:          "Suspicious Package",
					AnnotationType: AnnotationTypeBug,
					Summary:        fmt.Sprintf("Package %s@%s is suspicious", pkg.GetPackage().GetName(), pkg.GetPackage().GetVersion()),
					Severity:       AnnotationSeverityHigh,
					FilePath:       manifestPath,
					Link:           threatLink,
				})
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

		for _, violation := range pkg.GetViolations() {
			annotations = append(annotations, CodeInsightsAnnotation{
				Title:          violation.GetFilter().GetSummary(),
				AnnotationType: AnnotationTypeCodeSmell,
				Summary:        violation.GetFilter().GetDescription(),
				Severity:       AnnotationSeverityMedium, // Default severity for violations
				FilePath:       manifestPath,
			})
		}

		for _, threat := range pkg.GetThreats() {
			annotations = append(annotations, CodeInsightsAnnotation{
				Title:          "Threat Detected",
				AnnotationType: AnnotationTypeBug,
				Summary:        fmt.Sprintf("Threat ID: %s", threat.GetId()),
				Severity:       AnnotationSeverityHigh, // Default severity for threats
				FilePath:       manifestPath,
			})
		}
	}

	return annotations, nil
}

func createNumericCodeInsightsDataPoint(title string, value int) CodeInsightsData {
	return CodeInsightsData{
		Title: title,
		Type:  DataTypeNumber,
		Value: value,
	}
}
