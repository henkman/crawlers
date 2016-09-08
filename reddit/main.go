package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	_subreddit string
)

func init() {
	flag.StringVar(&_subreddit, "s", "", "subreddit to spit out")
	flag.Parse()
}

func main() {
	if _subreddit == "" {
		flag.Usage()
		return
	}
	requestDoc := func(c http.Client, method, url string, body io.Reader) (*goquery.Document, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent",
			"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.94 Safari/537.36")
		res, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		return goquery.NewDocumentFromResponse(res)
	}
	var c http.Client
	{
		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Fatal(err)
		}
		c.Jar = jar
	}
	var doc *goquery.Document
	{
		tdoc, err := requestDoc(c, "GET", _subreddit, nil)
		if err != nil {
			log.Fatal(err)
		}
		doc = tdoc
	}
	for {
		doc.Find(".thing[data-url]").Each(func(i int, s *goquery.Selection) {
			if url, ok := s.Attr("data-url"); ok {
				fmt.Println(url)
			}
		})
		next := doc.Find(".next-button a")
		if next.Length() == 0 {
			break
		}
		a, ok := next.Attr("href")
		if !ok {
			break
		}
		{
			tdoc, err := requestDoc(c, "GET", a, nil)
			if err != nil {
				log.Fatal(err)
			}
			doc = tdoc
		}
		time.Sleep(time.Second)
	}
}
