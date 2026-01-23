package main

import (
	"bytes"
	"flag"
	"os"

	"github.com/safedep/dry/api/pb"
	jsonreportspec "github.com/safedep/vet/gen/jsonreport"
)

func main() {
	var jsonReportFile string

	flag.StringVar(&jsonReportFile, "json-report-file", "", "Vet generated JSON report file path")
	flag.Parse()

	if jsonReportFile == "" {
		os.Exit(1)
	}

	data, err := os.ReadFile(jsonReportFile)
	if err != nil {
		os.Exit(1)
	}

	var report jsonreportspec.Report
	err = pb.FromJson(bytes.NewReader(data), &report)
}
