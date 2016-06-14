package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

var (
	_url string
)

func init() {
	flag.StringVar(&_url, "u", "", "url")
	flag.Parse()
}

func downloadThread(in chan string, wg *sync.WaitGroup) {
	for u := range in {
		doc, err := goquery.NewDocument(u)
		if err != nil {
			log.Println(err)
		}
		img := doc.Find("img.the-photo")
		if img.Length() == 0 {
			continue
		}
		src, ok := img.Attr("src")
		if !ok {
			continue
		}
		s := strings.LastIndex(src, "/")
		if s == -1 {
			continue
		}
		f := src[s+1:]
		if _, err := os.Stat(f); err == nil {
			continue
		}
		fd, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0750)
		if err != nil {
			log.Fatal(err)
			continue
		}
		r, err := http.Get(src)
		if err != nil {
			fd.Close()
			log.Fatal(err)
			continue
		}
		io.Copy(fd, r.Body)
		r.Body.Close()
		fd.Close()
	}
	wg.Done()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if _url == "" {
		flag.Usage()
		return
	}
	var wg sync.WaitGroup
	wg.Add(runtime.NumCPU() - 1)
	in := make(chan string)
	for i := 0; i < runtime.NumCPU()-1; i++ {
		go downloadThread(in, &wg)
	}
	{
		u := _url
		for {
			doc, err := goquery.NewDocument(u)
			if err != nil {
				log.Fatal(err)
			}
			doc.Find(".meta a").Each(func(i int, s *goquery.Selection) {
				if href, ok := s.Attr("href"); ok {
					in <- href
				}
			})
			next := doc.Find("a.next")
			if next.Length() == 0 {
				break
			}
			href, ok := next.Attr("href")
			if !ok {
				break
			}
			u = _url + href
		}
		close(in)
	}
	wg.Wait()
}