package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

var (
	_user string
)

func init() {
	flag.StringVar(&_user, "u", "", "user")
	flag.Parse()
}

func main() {
	if _user == "" {
		flag.Usage()
		return
	}
	getFile := func(file string, url string) error {
		if _, err := os.Stat(file); err == nil {
			log.Println(file, "already exists")
			return nil
		}
		log.Println("downloading", file)
		r, err := http.Get(url)
		if err != nil {
			return err
		}
		fd, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			r.Body.Close()
			return err
		}
		io.Copy(fd, r.Body)
		r.Body.Close()
		fd.Close()
		return nil
	}
	user := _user
	max_id := ""
	for {
		u := fmt.Sprintf("https://www.instagram.com/%s/media", user)
		if max_id != "" {
			u += fmt.Sprint("?max_id=", max_id)
		}
		r, err := http.Get(u)
		if err != nil {
			log.Fatal(err)
		}
		var ig struct {
			Items []struct {
				Images struct {
					StandardResolution struct {
						URL string `json:"url"`
					} `json:"standard_resolution"`
				} `json:"images"`
				Type   string `json:"type"`
				ID     string `json:"id"`
				Videos struct {
					StandardResolution struct {
						URL string `json:"url"`
					} `json:"standard_resolution"`
				} `json:"videos,omitempty"`
			} `json:"items"`
			MoreAvailable bool `json:"more_available"`
		}
		if err := json.NewDecoder(r.Body).Decode(&ig); err != nil {
			r.Body.Close()
			log.Fatal(err)
		}
		r.Body.Close()
		for _, item := range ig.Items {
			var url string
			if item.Type == "image" {
				url = item.Images.StandardResolution.URL
			} else if item.Type == "video" {
				url = item.Videos.StandardResolution.URL
			} else {
				continue
			}
			s := strings.LastIndex(url, "/")
			file := url[s+1:]
			q := strings.Index(url, "?")
			if q != -1 {
				file = url[s+1 : q]
			}
			getFile(file, url)
		}
		max_id = ig.Items[len(ig.Items)-1].ID
		if !ig.MoreAvailable {
			break
		}
	}
}