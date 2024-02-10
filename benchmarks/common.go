package tests

import (
	"fmt"
	"os"
)

var (
	// small, medium and large fixtures are from https://github.com/buger/jsonparser/blob/f04e003e4115787c6272636780bc206e5ffad6c4/benchmark/benchmark.go
	smallFixture  = getFromFile("../testdata/small.json")
	mediumFixture = getFromFile("../testdata/medium.json")
	largeFixture  = getFromFile("../testdata/large.json")

	// canada, citm and twitter fixtures are from https://github.com/serde-rs/json-benchmark/tree/0db02e043b3ae87dc5065e7acb8654c1f7670c43/data
	canadaFixture  = getFromFile("../testdata/canada.json")
	citmFixture    = getFromFile("../testdata/citm_catalog.json")
	twitterFixture = getFromFile("../testdata/twitter.json")
)

func getFromFile(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic(fmt.Errorf("cannot read %s: %s", filename, err))
	}
	return string(data)
}
