package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const endpoint string = "https://egov.uscis.gov/casestatus/mycasestatus.do?appReceiptNum="

type Receipt struct {
	AreaCode string
	Number   uint32
}

type Application struct {
	Receipt Receipt
	Status  string
	Date    time.Time
}

func isTarget(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "current-status-sec" {
				return true
			}
		}
	}
	return false
}

func isAppointmentSection(node *html.Node) bool {
	if node.Type == html.ElementNode && node.Data == "div" {
		for _, attr := range node.Attr {
			if attr.Key == "class" && attr.Val == "col-lg-12 appointment-sec center" {
				return true
			}
		}
	}
	return false
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

func (myApp *Application) parseDoc(node *html.Node) {
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
						const shortForm = "Jan 2, 2006"
						rDate, _ := time.Parse(shortForm, ss[0][3:]+strings.Trim(ss[1], ","))
						myApp.Date = rDate
					}
				}
			}
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		myApp.parseDoc(c)
	}
}

func main() {
	// example := "YSC1790219538"
	// resp, _ := http.Get(endpoint + example)
	myApp := &Application{
		Receipt: Receipt{
			AreaCode: "YSC",
			Number:   1790219538,
		},
	}
	file, _ := os.Open("example.html")
	fileReader := bufio.NewReader(file)
	doc, _ := html.Parse(fileReader)
	myApp.parseDoc(doc)
	fmt.Println(myApp)
}
