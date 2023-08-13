package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Specifications for getting chart data
const DefaultYFinanceProtocol = "https"
const DefaultYFinanceDomain = "query1.finance.yahoo.com"
const DefaultYFinanceEndpoint = "v8/finance/chart"
const DefaultYFinanceRegion = "US"
const DefaultYFinanceLang = "en-US"
const DefaultYFinanceIncludePrePost = "false"
const DefaultYFinanceInterval = "1m"
const DefaultYFinanceUseYfid = "true"
const DefaultYFinanceRange = "1d"
const DefaultYFinanceCorsDomain = "finance.yahoo.com"
const DefaultYFinanceTSRC = "finance"

// Specifications for getting bid-ask prices and volumes
const YFinanceBidAskDomain = "finance.yahoo.com"
const YFinanceBidAskTSRC = "fin-srch"
const YFinanceBidAskEndpoint = "quote"

// Configuration for sending requests
const DefaultYFinanceGetExpiration = 5000 * time.Millisecond

type Request struct {
	Scheme         string
	Host           string
	Path           string
	Region         string
	Lang           string
	IncludePrePost string
	Interval       string
	UseYfid        string
	Range          string
	CorsDomain     string
	TSRC           string
}

func InitDefaultRequest(ticker string) Request {
	return Request{
		Scheme:         DefaultYFinanceProtocol,
		Host:           DefaultYFinanceDomain,
		Path:           DefaultYFinanceEndpoint + "/" + ticker,
		Region:         DefaultYFinanceRegion,
		Lang:           DefaultYFinanceLang,
		IncludePrePost: DefaultYFinanceIncludePrePost,
		Interval:       DefaultYFinanceInterval,
		UseYfid:        DefaultYFinanceUseYfid,
		Range:          DefaultYFinanceRange,
		CorsDomain:     DefaultYFinanceCorsDomain,
		TSRC:           DefaultYFinanceTSRC,
	}
}

func InitBidAskRequest(ticker string) Request {
	return Request{
		Scheme: DefaultYFinanceProtocol,
		Host:   YFinanceBidAskDomain,
		Path:   YFinanceBidAskEndpoint + "/" + ticker,
		Region: DefaultYFinanceRegion,
		TSRC:   YFinanceBidAskTSRC,
	}
}

func (r *Request) String() string {
	u := strings.Builder{}
	u.WriteString(r.Scheme)
	u.WriteString("://")
	u.WriteString(r.Host)
	u.WriteString("/")
	u.WriteString(r.Path)
	u.WriteString("?")
	u.WriteString("region=")
	u.WriteString(r.Region)
	u.WriteString("&lang=")
	u.WriteString(r.Lang)
	u.WriteString("&includePrePost=")
	u.WriteString(r.IncludePrePost)
	u.WriteString("&interval=")
	u.WriteString(r.Interval)
	u.WriteString("&useYfid=")
	u.WriteString(string(r.UseYfid))
	u.WriteString("&range=")
	u.WriteString(r.Range)
	u.WriteString("&corsDomain=")
	u.WriteString(r.CorsDomain)
	u.WriteString("&.tsrc=")
	u.WriteString(r.TSRC)
	return u.String()
}

func (r *Request) GetQuote() (Quote, error) {
	var q Quote
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultYFinanceGetExpiration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", r.String(), nil)
	if err != nil {
		return q, err
	}
	resp, err := client.Do(req)

	if err != nil {
		return q, err
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return q, err
	}
	if err = json.Unmarshal(b, &q); err != nil {
		return q, err
	}
	return q, nil
}

// GetBidAsk requests the raw html and extracts the bid, ask volume and prices directly from the DOM.
func (r *Request) GetBidAsk() (BidAsk, error) {
	var ba BidAsk
	client := &http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), DefaultYFinanceGetExpiration)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", r.String(), nil)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("User-Agent", "")

	if err != nil {
		return ba, err
	}
	resp, err := client.Do(req)

	if err != nil {
		return ba, err
	}
	h, err := io.ReadAll(resp.Body)
	if err != nil {
		return ba, err
	}
	parsed, err := parseYFinanceHTML(string(h))
	if err != nil {
		return ba, err
	}
	b, err := extractByAttribute(&parsed, "td", "data-test", "BID-value")
	if err != nil {
		return ba, err
	}
	bidPriceString, bidVolumeString, err := parseRawBidAsk(b)
	if err != nil {
		return ba, err
	}
	ba.BidPrice, err = strconv.ParseFloat(bidPriceString, 64)
	if err != nil {
		return ba, err
	}
	ba.BidVolume, err = strconv.ParseInt(bidVolumeString, 10, 64)
	if err != nil {
		return ba, err
	}

	a, err := extractByAttribute(&parsed, "td", "data-test", "ASK-value")
	if err != nil {
		return ba, err
	}
	askPriceString, askVolumeString, err := parseRawBidAsk(a)
	if err != nil {
		return ba, err
	}
	ba.AskPrice, err = strconv.ParseFloat(askPriceString, 64)
	if err != nil {
		return ba, err
	}
	ba.AskVolume, err = strconv.ParseInt(askVolumeString, 10, 64)
	if err != nil {
		return ba, err
	}
	return ba, nil
}

func parseYFinanceHTML(htmlString string) (html.Node, error) {
	doc, err := html.Parse(strings.NewReader(htmlString))
	if err != nil {
		return html.Node{}, err
	}
	return *doc, nil
}

func extractByAttribute(doc *html.Node, tagName, attributeName, attributeValue string) (string, error) {
	if doc == nil {
		return "", errors.New("passed document pointer is nil")
	}
	if n := findNodeByAttribute(doc, tagName, attributeName, attributeValue); n != nil {
		return extractTextContent(n), nil
	}
	return "", errors.New("extracted node was empty")
}

// findNodeByAttribute returns the first matching node using DFS
func findNodeByAttribute(node *html.Node, tagName, attributeName, attributeValue string) *html.Node {
	if node.Type == html.ElementNode && node.Data == tagName {
		for _, attr := range node.Attr {
			if attr.Key == attributeName && attr.Val == attributeValue {
				return node
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if foundNode := findNodeByAttribute(child, tagName, attributeName, attributeValue); foundNode != nil {
			return foundNode
		}
	}
	return nil
}

func extractTextContent(node *html.Node) string {
	var content string
	if node.Type == html.TextNode {
		content = node.Data
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		content += extractTextContent(child)
	}
	return content
}

func parseRawBidAsk(raw string) (string, string, error) {
	parsed := strings.Split(raw, " x ")
	if len(parsed) != 2 {
		return "", "", errors.New(fmt.Sprintf("failed to split: %s", raw))
	}
	return parsed[0], parsed[1], nil
}
