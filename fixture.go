package main

import (
	"strings"
)

type native struct {
	Slug              string `yaml:"slug"`
	AffiliationSource string `yaml:"affiliation_source"`
}

type fixture struct {
	Disabled    bool         `yaml:"disabled"`
	AllowEmpty  bool         `yaml:"allow_empty"`
	Native      native       `yaml:"native"`
	DataSources []dataSource `yaml:"data_sources"`
	Aliases     []alias      `yaml:"aliases"`
	Slug        string       `yaml:"-"`
}

type dataSource struct {
	Slug        string `yaml:"slug"`
	IndexSuffix string `yaml:"index_suffix"`
	FullSlug    string `yaml:"-"`
}

type aliasView struct {
	Name string `yaml:"name"`
}

type alias struct {
	From  string      `yaml:"from"`
	To    []string    `yaml:"to"`
	Views []aliasView `yaml:"views"`
}

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
