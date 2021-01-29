package main

import (
	"bytes"
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v2"
)

var (
	gESURL        string
	gBranch       string
	gRecipients   string
	gHostname     string
	noDropPattern = regexp.MustCompile(`^(.+-f-.+|.+-earned_media|.+-slack)$`)
	gHTML         = "<!DOCTYPE html>\n<html>\n<head>\n  <meta charset=\"utf-8\">\n  <title>%s</title>\n</head>\n<body>\n%s\n</body>\n</html>\n"
)

type esIndex struct {
	Index string `json:"index"`
}

type esAlias struct {
	Alias string `json:"alias"`
	Index string `json:"index"`
}

func processIndexesInfo(fixtures []fixture) (info string) {
	should := make(map[string]struct{})
	fromFull := make(map[string]string)
	toFull := make(map[string]string)
	for _, fixture := range fixtures {
		slug := fixture.Slug
		slug = strings.Replace(slug, "/", "-", -1)
		for _, ds := range fixture.DataSources {
			if ds.Slug == "earned_media" {
				continue
			}
			// Skip configured but empty data sources
			if len(ds.Endpoints) == 0 && len(ds.Projects) == 0 {
				continue
			}
			idxSlug := "sds-" + slug + "-" + ds.FullSlug
			idxSlug = strings.Replace(idxSlug, "/", "-", -1)
			if idxSlug == "" || idxSlug == "sds-" {
				continue
			}
			should[idxSlug] = struct{}{}
			if ds.Slug != ds.FullSlug {
				idx := "sds-" + slug + "-" + ds.Slug
				idx = strings.Replace(idx, "/", "-", -1)
				fromFull[idxSlug] = idx
				toFull[idx] = idxSlug
			}
		}
		for _, alias := range fixture.Aliases {
			idxSlug := alias.From
			if strings.HasPrefix(alias.From, "pattern:") || strings.HasPrefix(alias.From, "bitergia-") || strings.HasSuffix(alias.From, "-raw") {
				continue
			}
			if idxSlug == "" || idxSlug == "sds-" {
				continue
			}
			should[idxSlug] = struct{}{}
		}
	}
	method := "GET"
	url := fmt.Sprintf("%s/_cat/indices?format=json", gESURL)
	rurl := "/_cat/indices?format=json"
	req, err := http.NewRequest(method, os.ExpandEnv(url), nil)
	if err != nil {
		fatalf("new request error: %+v for %s url: %s\n", err, method, rurl)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fatalf("do request error: %+v for %s url: %s\n", err, method, rurl)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fatalf("readAll request error: %+v for %s url: %s\n", err, method, rurl)
			return
		}
		fatalf("method:%s url:%s status:%d\n%s\n", method, rurl, resp.StatusCode, body)
		return
	}
	indices := []esIndex{}
	err = jsoniter.NewDecoder(resp.Body).Decode(&indices)
	if err != nil {
		fatalf("JSON decode error: %+v for %s url: %s\n", err, method, rurl)
		return
	}
	got := make(map[string]struct{})
	for _, index := range indices {
		sIndex := index.Index
		if !strings.HasPrefix(sIndex, "sds-") || strings.HasSuffix(sIndex, "-raw") {
			continue
		}
		got[sIndex] = struct{}{}
	}
	missing := []string{}
	extra := []string{}
	rename := make(map[string]string)
	for fullIndex := range should {
		_, ok := got[fullIndex]
		if !ok {
			index := fromFull[fullIndex]
			_, ok := got[index]
			if ok {
				rename[index] = fullIndex
			} else {
				missing = append(missing, fullIndex)
			}
		}
	}
	renames := []string{}
	for index := range got {
		_, ok := should[index]
		if !ok {
			fullIndex, ok := rename[index]
			if !ok {
				extra = append(extra, index)
			} else {
				renames = append(renames, index+" -> "+fullIndex)
			}
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 {
		info += fmt.Sprintf("<b><p style=\"color:red\">missing %d indices:</p></b> <small>%s</small>\n", len(missing), html.EscapeString(strings.Join(missing, ", ")))
		fmt.Printf("missing %d indices: %s\n", len(missing), strings.Join(missing, ", "))
	}
	newExtra := []string{}
	for _, idx := range extra {
		if noDropPattern.MatchString(idx) {
			continue
		}
		newExtra = append(newExtra, idx)
	}
	extra = newExtra
	if len(extra) > 0 {
		if info != "" {
			info += "\n"
		}
		info += fmt.Sprintf("<b><p style=\"color:blue\">following %d indices should be removed:</p></b> <small>%s</small>\n", len(extra), html.EscapeString(strings.Join(extra, ", ")))
		fmt.Printf("following %d indices should be removed: %s\n", len(extra), strings.Join(extra, ", "))
	}
	if len(renames) > 0 {
		if info != "" {
			info += "\n"
		}
		info += fmt.Sprintf("<b><p style=\"color:blue\">%d indices should be renamed:</p></b> <small>%s</small>\n\n", len(renames), html.EscapeString(strings.Join(renames, ", ")))
		fmt.Printf("%d indices should be renamed: %s\n", len(renames), strings.Join(renames, ", "))
	}
	return
}

func dropUnusedAliasesInfo(fixtures []fixture) (info string) {
	should := make(map[string]struct{})
	for _, fixture := range fixtures {
		for _, alias := range fixture.Aliases {
			for _, to := range alias.To {
				if strings.HasSuffix(to, "-raw") {
					continue
				}
				should[strings.Replace(to, "/", "-", -1)] = struct{}{}
			}
			for _, view := range alias.Views {
				if strings.HasSuffix(view.Name, "-raw") {
					continue
				}
				should[strings.Replace(view.Name, "/", "-", -1)] = struct{}{}
			}
		}
	}
	method := "GET"
	url := fmt.Sprintf("%s/_cat/aliases?format=json", gESURL)
	rurl := "/_cat/aliases?format=json"
	req, err := http.NewRequest(method, os.ExpandEnv(url), nil)
	if err != nil {
		fatalf("new request error: %+v for %s url: %s\n", err, method, rurl)
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fatalf("do request error: %+v for %s url: %s\n", err, method, rurl)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fatalf("readAll request error: %+v for %s url: %s\n", err, method, rurl)
			return
		}
		fatalf("nethod:%s url:%s status:%d\n%s\n", method, rurl, resp.StatusCode, body)
		return
	}
	aliases := []esAlias{}
	err = jsoniter.NewDecoder(resp.Body).Decode(&aliases)
	if err != nil {
		fatalf("JSON decode error: %+v for %s url: %s\n", err, method, rurl)
		return
	}
	got := make(map[string]struct{})
	for _, alias := range aliases {
		sAlias := alias.Alias
		if !strings.HasPrefix(sAlias, "sds-") || strings.HasSuffix(sAlias, "-raw") {
			continue
		}
		got[sAlias] = struct{}{}
	}
	missing := []string{}
	extra := []string{}
	for alias := range should {
		_, ok := got[alias]
		if !ok {
			missing = append(missing, alias)
		}
	}
	for alias := range got {
		_, ok := should[alias]
		if !ok {
			extra = append(extra, alias)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)
	if len(missing) > 0 {
		info += fmt.Sprintf("<b><p style=\"color:red\">missing %d aliases:</p></b> <small>%s</small>\n", len(missing), html.EscapeString(strings.Join(missing, ", ")))
		fmt.Printf("missing %d aliases: %s\n", len(missing), strings.Join(missing, ", "))
	}
	newExtra := []string{}
	for _, idx := range extra {
		if noDropPattern.MatchString(idx) {
			continue
		}
		newExtra = append(newExtra, idx)
	}
	extra = newExtra
	if len(extra) > 0 {
		if info != "" {
			info += "\n"
		}
		info += fmt.Sprintf("<b><p style=\"color:blue\">%d aliases to delete:</p></b> <small>%s</small>\n", len(extra), html.EscapeString(strings.Join(extra, ", ")))
		fmt.Printf("%d aliases to delete: %s\n", len(extra), strings.Join(extra, ", "))
	}
	return
}

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
	for i, dataSource := range fx.DataSources {
		fs := dataSource.Slug + dataSource.IndexSuffix
		fs = strings.Replace(fs, "/", "-", -1)
		fx.DataSources[i].FullSlug = fs
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
		if idxSlug == "" || idxSlug == "sds-" {
			continue
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

func processFixtureFiles(fixtureFiles []string) string {
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
	idxInfo := processIndexesInfo(fixtures)
	aliasInfo := dropUnusedAliasesInfo(fixtures)
	finalInfo := ""
	if idxInfo != "" {
		finalInfo = "<b>Indices status (" + gBranch + " environment):\n==================================</b>\n" + idxInfo
		finalInfo += "<b>==================================</b>\n"
	}
	if aliasInfo != "" {
		if finalInfo != "" {
			finalInfo += "\n\n"
		}
		finalInfo += "<b>Aliases status (" + gBranch + " environment):\n==================================</b>\n" + aliasInfo
		finalInfo += "<b>==================================</b>\n"
	}
	return finalInfo
}

func sendStatusEmail(body string) error {
	fmt.Printf("sending email(s) to %s\n", gRecipients)
	title := "ES " + gBranch + " monitor status"
	htmlBody := fmt.Sprintf(gHTML, title, strings.Replace(body, "\n", "<br/>\n", -1))
	ary := strings.Split(gRecipients, ",")
	for _, recipient := range ary {
		recipient = strings.TrimSpace(recipient)
		//fmt.Printf("sending email to %s\n", recipient)
		data := fmt.Sprintf(
			"From: ES-monitor@%s\n"+
				"To: %s\n"+
				"Subject: %s\n"+
				"Content-Type: text/html\n"+
				"MIME-Version: 1.0\n"+
				"\n"+
				"%s\n",
			gHostname,
			recipient,
			title,
			htmlBody,
		)
		res, err := execCommandWithStdin([]string{"sendmail", recipient}, bytes.NewBuffer([]byte(data)))
		if err != nil {
			fmt.Printf("Error sending email to %s: %+v\n%s\n", recipient, err, res)
		}
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fatalf("%s: you must secify path to fixtures diectory as an argument\n", os.Args[0])
	}
	gESURL = os.Getenv("ES_URL")
	if gESURL == "" {
		fatalf("%s: you must set ES_URL env variable\n", os.Args[0])
	}
	gBranch = os.Getenv("BRANCH")
	if gBranch == "" {
		fatalf("%s: you must set BRANCH env variable\n", os.Args[0])
	}
	gRecipients = os.Getenv("RECIPIENTS")
	if gRecipients == "" {
		fatalf("%s: you must set RECIPIENTS env variable\n", os.Args[0])
	}
	gHostname, _ = os.Hostname()
	fixtures := getFixtures(os.Args[1])
	info := processFixtureFiles(fixtures)
	if info != "" {
		//fmt.Printf("%s\n", info)
		fatalOnError(sendStatusEmail(info))
	}
	fmt.Printf("finished\n")
}
