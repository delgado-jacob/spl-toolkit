package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/delgado-jacob/spl-toolkit/pkg/corpus"
	"github.com/delgado-jacob/spl-toolkit/pkg/graph"
)

func main() {
	request, err := corpus.DecodeRequest([]byte(`{"schema_version":1,"documents":[{"id":"hosts","document":{"text":"search host=web | table host"}}]}`))
	if err != nil {
		log.Fatal(err)
	}
	report, err := corpus.Scan(request)
	if err != nil {
		log.Fatal(err)
	}
	result, err := graph.Export(report)
	if err != nil {
		log.Fatal(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		log.Fatal(err)
	}
}
