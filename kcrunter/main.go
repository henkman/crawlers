package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: kcrunter url")
		return
	}
	url := os.Args[1]
	reUrl := regexp.MustCompile("krautchan.net/([^/]+)/thread-\\d+\\.html")
	m := reUrl.FindStringSubmatch(url)
	if m == nil {
		fmt.Println("not a valid url")
		return
	}
	doc, err := goquery.NewDocument(url)
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	n := 1
	if runtime.NumCPU() > n {
		n = runtime.NumCPU() - 1
	}
	type Request struct {
		Url  string
		File string
	}
	dlchan := make(chan Request)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			for dlreq := range dlchan {
				if req, err := http.Get(dlreq.Url); err == nil {
					if fd, err := os.OpenFile(dlreq.File, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0750); err == nil {
						io.Copy(fd, req.Body)
						fd.Close()
					}
					req.Body.Close()
				}
			}
			wg.Done()
		}()
	}
	doc.Find(".filename a").Each(func(i int, s *goquery.Selection) {
		if href, ok := s.Attr("href"); ok {
			file := filepath.Base(href)
			if _, err := os.Stat(file); os.IsNotExist(err) {
				fmt.Println("downloading", file)
				dlchan <- Request{
					"https://krautchan.net" + href,
					file,
				}
			} else {
				fmt.Println(file, "already exists")
			}
		}
	})
	close(dlchan)
	wg.Wait()
}
