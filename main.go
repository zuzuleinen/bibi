package main

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/mailgun/mailgun-go"
	"github.com/yhat/scrape"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"log"
	"math/rand"
	"net/http"
	"os"
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

type Configuration struct {
	MailGunDomain     string
	MailGunPublicKey  string
	MailGunPrivateKey string
	SenderEmail       string
	RecipientEmail    string
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

	c := random(cogs)
	c.display()

	if len(os.Args) > 1 && os.Args[1] == "--email" {
		send(c, fetchConfig())
	}
}

//Construct a cognitive bias from a table row
func makeCognitiveBias(row *html.Node) cognitiveBias {
	cells := scrape.FindAllNested(row, scrape.ByTag(atom.Td))

	if len(cells) < 2 {
		return cognitiveBias{}
	}

	firstCell := cells[0]
	secondCell := cells[1]

	urlNode, ok := scrape.Find(firstCell, scrape.ByTag(atom.A))

	url := ""
	if ok {
		url = wikiPrefixUrl + scrape.Attr(urlNode, "href")
	}
	return cognitiveBias{scrape.Text(firstCell), scrape.Text(secondCell), url}
}

//Display a cognitive bias to the terminal
func (c *cognitiveBias) display() {
	color.Green(c.name)
	color.White(c.description)

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

//Fetch config values from conf.json
func fetchConfig() Configuration {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	configuration := Configuration{}

	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}
	return configuration
}

//Send an e-mail with the cognitive bias
func send(c *cognitiveBias, conf Configuration) {
	mg := mailgun.NewMailgun(
		conf.MailGunDomain,
		conf.MailGunPrivateKey,
		conf.MailGunPublicKey,
	)

	subject := "New Bias: " + c.name
	body := "Hey, here is your daily dose of cognitive biases provided by <strong>bibi</strong>: <br /><br />"
	body += fmt.Sprintf("<strong>%s</strong><br />", c.name)
	body += fmt.Sprintf("%s <br />", c.description)
	if c.url != "" {
		body += fmt.Sprintf("Find out more: <a href=\"%s\">Wikipedia Link<a/>", c.url)
	}

	m := mg.NewMessage(conf.SenderEmail, subject, "", conf.RecipientEmail)
	m.SetHtml(body)

	_, _, err := mg.Send(m)

	if err != nil {
		log.Fatal(err)
	}
}