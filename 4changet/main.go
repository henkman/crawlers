package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	_url string
)

func init() {
	flag.StringVar(&_url, "u", "", "url")
	flag.Parse()
}

func main() {
	if _url == "" {
		flag.Usage()
		return
	}
	doc, err := goquery.NewDocument(_url)
	if err != nil {
		log.Panicln(err)
	}
	var wg sync.WaitGroup
	doc.Find(".fileText a").Each(func(i int, s *goquery.Selection) {
		fmt.Println(s)
		if href, ok := s.Attr("href"); ok {
			u := "http:" + href
			name := s.Text()
			path := "./" + name
			if _, err := os.Stat(path); err == nil {
				fmt.Println(name, "already exists")
				return
			}
			fmt.Println("downloading", name)
			time.Sleep(time.Millisecond * 100)
			wg.Add(1)
			go func() {
				if fd, err := os.OpenFile("./"+name, os.O_WRONLY|os.O_CREATE, 0600); err == nil {
					if r, err := http.Get(u); err == nil {
						io.Copy(fd, r.Body)
						r.Body.Close()
					}
					fd.Close()
				}
				wg.Done()
			}()
		}
	})
	wg.Wait()
}