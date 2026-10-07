package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/workflow"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("usage: go run ./examples/go/workflow REQUEST.json")
	}
	raw, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	report, err := workflow.AssessJSON(raw)
	if err != nil {
		log.Fatal(err)
	}
	evidence, err := workflow.Evidence(workflow.EvidenceRequest{SchemaVersion: 1, Report: report, Include: []string{}})
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(evidence); err != nil {
		log.Fatal(err)
	}
}
