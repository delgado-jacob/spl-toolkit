package main

import (
	"fmt"
	"log"

	"github.com/delgado-jacob/spl-toolkit/pkg/mapper"
)

func main() {
	m := mapper.New()
	if err := m.LoadMappings([]byte(`[{"source":"src_ip","target":"source_ip"}]`)); err != nil {
		log.Fatal(err)
	}
	mapped, err := m.MapQuery("search src_ip=1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(mapped)
}
