package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	BASEID   int = 1790100000
	MAXITER  int = 0
	ENDPOINT     = "https://egov.uscis.gov/casestatus/mycasestatus.do?appReceiptNum="
)

type Receipt struct {
	AreaCode string
	Number   string
}

type Application struct {
	Receipt Receipt
	Status  string
	Date    string
}

func (r *Receipt) ToString() string {
	return r.AreaCode + r.Number
}

type Query struct {
	NodeData string
	NodeType html.NodeType
	Attr     *html.Attribute
}

// Check if a node has the wanted information
func foundNode(node *html.Node, q *Query) bool {
	if node.Type == q.NodeType && node.Data == q.NodeData {
		for _, attr := range node.Attr {
			if attr.Key == q.Attr.Key && attr.Val == q.Attr.Val {
				return true
			}
		}
	}
	return false
}

func (myApp *Application) parseDoc(node *html.Node, ch chan *Application) {
	found := false
	if node.Type == html.ElementNode {
		queryStatusParent := &Query{
			NodeType: html.ElementNode,
			NodeData: "div",
			Attr: &html.Attribute{
				Key: "class",
				Val: "col-lg-12 appointment-sec center",
			},
		}
		queryStatus := &Query{
			NodeType: html.ElementNode,
			NodeData: "div",
			Attr: &html.Attribute{
				Key: "class",
				Val: "rows text-center",
			},
		}

		if foundNode(node.Parent, queryStatusParent) && foundNode(node, queryStatus) {
			i := 0
			for iterNode := node.FirstChild; iterNode != nil; iterNode = iterNode.NextSibling {
				if iterNode.FirstChild != nil {
					i++
					text := string(iterNode.FirstChild.Data)
					if i == 1 {
						myApp.Status = text
					} else if i == 2 {
						ss := strings.SplitAfterN(text, ",", 3)
						rDate := ss[0][3:] + strings.Trim(ss[1], ",")
						myApp.Date = rDate
					}
					found = true
				}
			}
		}
	}
	if found {
		ch <- myApp
		return
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		myApp.parseDoc(c, ch)
	}
}

func main() {
	records := [][]string{
		{"case_id", "case_status", "date"},
	}
	ch := make(chan *Application)
	go func() {
		for id := 0; id <= MAXITER; id++ {
			nextId := BASEID + int(1000*id)
			myApp := &Application{
				Receipt: Receipt{
					AreaCode: "YSC",
					Number:   strconv.Itoa(nextId),
				},
			}
			resp, _ := http.Get(ENDPOINT + myApp.Receipt.ToString())
			doc, _ := html.Parse(resp.Body)
			myApp.parseDoc(doc, ch)
		}
		close(ch)
	}()

	for a := range ch {
		records = append(records, []string{a.Receipt.ToString(), a.Status, a.Date})
	}

	fileName := time.Now().String()
	csvFile, err := os.Create(fileName + ".csv")
	if err != nil {
		log.Fatal("Error open file")
	}

	defer csvFile.Close()
	w := csv.NewWriter(csvFile)
	w.WriteAll(records)
	if err := w.Error(); err != nil {
		log.Fatalln("error writing to cvs")
	}
}
