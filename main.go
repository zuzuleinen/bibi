package main

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"math/rand"
	"net/http"
	"time"
)

const (
	fetchUrl      = "https://en.wikipedia.org/wiki/List_of_cognitive_biases"
	wikiPrefixUrl = "https://en.wikipedia.org"
	tableClass    = "wikitable"
)

type cognitiveBias struct {
	name        string
	description string
	url         string
}

func main() {
	resp, err := http.Get(fetchUrl)
	if err != nil {
		panic(err)
	}
	root, err := html.Parse(resp.Body)
	if err != nil {
		panic(err)
	}

	tables := scrape.FindAll(root, scrape.ByClass(tableClass))
	var cogs []cognitiveBias

	for _, table := range tables {
		rows := scrape.FindAllNested(table, scrape.ByTag(atom.Tr))
		for k, row := range rows {
			if k == 0 {
				continue
			}
			cogs = append(cogs, makeCognitiveBias(row))
		}
	}

	g := random(cogs)
	g.display()
}

//Construct a cognitive bias from a table row
func makeCognitiveBias(row *html.Node) cognitiveBias {
	cells := scrape.FindAllNested(row, scrape.ByTag(atom.Td))

	if len(cells) > 0 {
		firstCell := cells[0]
		secondCell := cells[1]

		urlNode, ok := scrape.Find(firstCell, scrape.ByTag(atom.A))

		url := ""
		if ok {
			url = wikiPrefixUrl + scrape.Attr(urlNode, "href")
		}
		return cognitiveBias{scrape.Text(firstCell), scrape.Text(secondCell), url}
	}
	return cognitiveBias{}
}

func (c *cognitiveBias) display() {
	color.Green(c.name)
	fmt.Println(c.description)

	if c.url != "" {
		color.Yellow("Find out more: %s", c.url)
	}
}

//Fetch a random cognitive bias
func random(cogs []cognitiveBias) *cognitiveBias {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)
	return &cogs[r.Intn(len(cogs))]
}
