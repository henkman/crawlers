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
)

var (
	_url string
)

func init() {
	flag.StringVar(&_url, "u", "", "album url")
	flag.Parse()
}

func main() {
	if _url == "" {
		flag.Usage()
		return
	}
	var id string
	{
		reAlbum := regexp.MustCompile("imgur.com/a/([0-9a-zA-Z]+)$")
		m := reAlbum.FindStringSubmatch(_url)
		if m == nil {
			fmt.Println("url is not a valid album")
			return
		}
		id = m[1]
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
	{ // get cookies
		res, err := request(c, "GET", _url, nil)
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
}
