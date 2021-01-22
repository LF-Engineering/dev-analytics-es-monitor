package main

import (
	"strings"
)

func getFixtures(path string) (fixtures []string) {
	res, err := execCommand(
		[]string{
			"find",
			path,
			"-type",
			"f",
			"-iname",
			"*.y*ml",
		},
	)
	if err != nil {
		fatalf("Error finding fixtures: %+v\n", err)
	}
	fixtures = strings.Split(res, "\n")
	return
}
