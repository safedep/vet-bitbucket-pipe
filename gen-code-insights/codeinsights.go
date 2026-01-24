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

func (ci codeInsightsGenerator) GenerateReport() (*CodeInsightsReport, error) {
	vulnerabilities := 0
	malicious := 0
	suspicious := 0

	for _, pkg := range ci.reportData.GetPackages() {
		if len(pkg.GetVulnerabilities()) > 0 {
			vulnerabilities += len(pkg.GetVulnerabilities())
		}

		for _, m := range pkg.GetMalwareInfo() {
			switch m.GetType() {
			case jsonreportspec.MalwareType_SUSPICIOUS:
				suspicious++
			case jsonreportspec.MalwareType_MALICIOUS:
				malicious++
			}
		}
	}

	report := &CodeInsightsReport{
		Title:      ci.config.ReportTitle,
		Reporter:   ci.config.ReportVendor,
		ReportType: ReportTypeSecurity,
		Result:     ReportResultPassed,
	}

	// OPNIONATED
	// If anything exists, vulnerability or suspicious or malicious packages
	// then the PR is not "Safe" to Merge
	if (vulnerabilities + suspicious + malicious) > 0 {
		report.Result = ReportResultFailed
	}

	report.Details = fmt.Sprintf("Found %d vulnerabilities, %d suspcious packages and %d malicious packages.", vulnerabilities, suspicious, malicious)

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
