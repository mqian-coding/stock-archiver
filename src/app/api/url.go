package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

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
