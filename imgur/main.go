package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

var (
	_album string
	_image string
)

func init() {
	flag.StringVar(&_album, "a", "", "album url")
	flag.StringVar(&_image, "i", "", "image url")
	flag.Parse()
}

func main() {
	if _album == "" && _image == "" {
		flag.Usage()
		return
	}
	var c http.Client
	{
		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Fatal(err)
		}
		c.Jar = jar
	}
	request := func(c http.Client, method, url string, body io.Reader) (*http.Response, error) {
		req, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, err
		}
		req.Header.Set("User-Agent",
			"Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.94 Safari/537.36")
		return c.Do(req)
	}
	if _album != "" {
		var id string
		{
			reAlbum := regexp.MustCompile("imgur.com/a/([0-9a-zA-Z]+)$")
			m := reAlbum.FindStringSubmatch(_album)
			if m == nil {
				fmt.Println("url is not a valid album")
				return
			}
			id = m[1]
		}
		{ // get cookies
			res, err := request(c, "GET", _album, nil)
			if err != nil {
				log.Fatal(err)
			}
			res.Body.Close()
		}
		res, err := request(c,
			"GET",
			fmt.Sprintf("http://imgur.com/ajaxalbums/getimages/%s/hit.json", id),
			nil)
		if err != nil {
			log.Fatal(err)
		}
		var data struct {
			Data struct {
				Images []struct {
					Hash string `json:"hash"`
					Ext  string `json:"ext"`
				} `json:"images"`
			} `json:"data"`
		}
		if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
			res.Body.Close()
			log.Fatal(err)
		}
		res.Body.Close()
		for _, img := range data.Data.Images {
			fmt.Printf("http://i.imgur.com/%s%s\n", img.Hash, img.Ext)
		}
	} else {
		res, err := request(c, "GET", _image, nil)
		if err != nil {
			log.Fatal(err)
		}
		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			log.Fatal(err)
		}
		a := doc.Find(".post-image a")
		if url, ok := a.Attr("href"); ok {
			fmt.Println(url)
		}
	}
}
