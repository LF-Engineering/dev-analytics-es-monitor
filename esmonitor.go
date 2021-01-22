package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"gopkg.in/yaml.v2"
)

func processFixtureFile(ch chan fixture, fixtureFile string) (fx fixture) {
	defer func() {
		ch <- fx
	}()
	data, err := ioutil.ReadFile(fixtureFile)
	fatalOnError(err)
	fatalOnError(yaml.Unmarshal(data, &fx))
	slug := fx.Native.Slug
	if slug == "" {
		fatalf("Fixture file %s 'native' property has no 'slug' property (or is empty)\n", fixtureFile)
	}
	fx.Slug = slug
	if fx.Disabled == true {
		return
	}
	for ai, alias := range fx.Aliases {
		var idxSlug string
		if strings.HasPrefix(alias.From, "bitergia-") || strings.HasPrefix(alias.From, "pattern:") || strings.HasPrefix(alias.From, "postprocess-") {
			idxSlug = alias.From
		} else {
			idxSlug = "sds-" + alias.From
		}
		if !strings.HasPrefix(alias.From, "pattern:") {
			idxSlug = strings.Replace(idxSlug, "/", "-", -1)
		}
		fx.Aliases[ai].From = idxSlug
		for ti, to := range alias.To {
			idxSlug := ""
			if strings.HasPrefix(to, "postprocess") {
				idxSlug = "postprocess-sds-" + strings.TrimPrefix(to, "postprocess/")
			} else {
				idxSlug = "sds-" + to
			}
			idxSlug = strings.Replace(idxSlug, "/", "-", -1)
			fx.Aliases[ai].To[ti] = idxSlug

		}
		for vi, v := range alias.Views {
			idxSlug := "sds-" + v.Name
			idxSlug = strings.Replace(idxSlug, "/", "-", -1)
			fx.Aliases[ai].Views[vi].Name = idxSlug
		}
	}
	return
}

func processFixtureFiles(fixtureFiles []string) {
	thrN := runtime.NumCPU()
	runtime.GOMAXPROCS(thrN)
	ch := make(chan fixture)
	fixtures := []fixture{}
	nThreads := 0
	for _, fixtureFile := range fixtureFiles {
		if fixtureFile == "" {
			continue
		}
		go processFixtureFile(ch, fixtureFile)
		nThreads++
		if nThreads == thrN {
			fixture := <-ch
			nThreads--
			if fixture.Disabled != true {
				fixtures = append(fixtures, fixture)
			}
		}
	}
	for nThreads > 0 {
		fixture := <-ch
		nThreads--
		if fixture.Disabled != true {
			fixtures = append(fixtures, fixture)
		}
	}
	if len(fixtures) == 0 {
		fatalf("No fixtures read, this is error, please define at least one")
	}
	st := make(map[string]fixture)
	for _, fixture := range fixtures {
		slug := fixture.Native.Slug
		slug = strings.Replace(slug, "/", "-", -1)
		fixture2, ok := st[slug]
		if ok {
			fatalf("Duplicate slug %s in fixtures: %+v and %+v\n", slug, fixture, fixture2)
		}
		st[slug] = fixture
	}
	//processIndexes(fixtures)
	//dropUnusedAliases(fixtures)
	//processAliases(fixtures)
}

func main() {
	if len(os.Args) < 2 {
		fatalf("%s: you must secify path to fixtures diectory as an argument\n", os.Args[0])
	}
	fixtures := getFixtures(os.Args[1])
	processFixtureFiles(fixtures)
}
