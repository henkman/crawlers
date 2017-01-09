package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
)

var (
	reSharedData = regexp.MustCompile("window._sharedData = (.*?);</script>")
	_user        string
	_videos      bool
)

func init() {
	flag.BoolVar(&_videos, "v", false, "download videos")
	flag.StringVar(&_user, "u", "", "user")
	flag.Parse()
}

func getFile(cli *http.Client, url, file string) error {
	if _, err := os.Stat(file); err == nil {
		log.Println(file, "already exists")
		return nil
	}
	log.Println("downloading", file)
	r, err := cli.Get(url)
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

func getMedia(cli *http.Client, detailpage string) error {
	var mediaurl string
	{
		r, err := cli.Get(detailpage)
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			r.Body.Close()
			return err
		}
		r.Body.Close()
		m := reSharedData.FindStringSubmatch(string(data))
		if m == nil {
			return errors.New("sharedData not found")
		}
		b := bytes.NewBufferString(m[1])
		var sharedData struct {
			EntryData struct {
				PostPage []struct {
					Media struct {
						DisplaySrc string `json:"display_src"`
						IsVideo    bool   `json:"is_video"`
						VideoURL   string `json:"video_url"`
					} `json:"media"`
				} `json:"PostPage"`
			} `json:"entry_data"`
		}
		if err := json.NewDecoder(b).Decode(&sharedData); err != nil {
			return err
		}
		media := sharedData.EntryData.PostPage[0].Media
		if media.IsVideo {
			if !_videos {
				return nil
			}
			mediaurl = media.VideoURL
		} else {
			mediaurl = media.DisplaySrc
		}
	}
	var file string
	{
		u, err := url.Parse(mediaurl)
		if err != nil {
			return err
		}
		s := strings.LastIndex(u.Path, "/")
		if s == -1 {
			return errors.New("img url is fucked")
		}
		file = u.Path[s+1:]
	}
	return getFile(cli, mediaurl, file)
}

func main() {
	if _user == "" {
		flag.Usage()
		return
	}
	var cli http.Client
	{
		jar, err := cookiejar.New(nil)
		if err != nil {
			log.Fatal(err)
		}
		cli = http.Client{Jar: jar}
	}
	var userid string
	var lastid string
	var hasNext bool
	{
		var sharedData struct {
			EntryData struct {
				ProfilePage []struct {
					User struct {
						ID    string `json:"id"`
						Media struct {
							PageInfo struct {
								HasNextPage bool `json:"has_next_page"`
							} `json:"page_info"`
							Nodes []struct {
								Code string `json:"code"`
								ID   string `json:"id"`
							} `json:"nodes"`
						} `json:"media"`
					} `json:"user"`
				} `json:"ProfilePage"`
			} `json:"entry_data"`
		}
		{
			r, err := cli.Get(
				fmt.Sprintf("https://www.instagram.com/%s/", _user))
			if err != nil {
				log.Fatal(err)
			}
			data, err := ioutil.ReadAll(r.Body)
			if err != nil {
				r.Body.Close()
				log.Fatal(err)
			}
			r.Body.Close()
			m := reSharedData.FindStringSubmatch(string(data))
			if m == nil {
				log.Fatal("sharedData not found")
			}
			b := bytes.NewBufferString(m[1])
			if err := json.NewDecoder(b).Decode(&sharedData); err != nil {
				log.Fatal(err)
			}
			if len(sharedData.EntryData.ProfilePage) == 0 {
				log.Fatal("sharedData has no profile page")
			}
		}
		u := sharedData.EntryData.ProfilePage[0].User
		for _, m := range u.Media.Nodes {
			getMedia(&cli, fmt.Sprintf(
				"https://www.instagram.com/p/%s/",
				m.Code))
			lastid = m.ID
		}
		userid = u.ID
		hasNext = u.Media.PageInfo.HasNextPage
	}
	for hasNext {
		ps := url.Values{
			"q": []string{fmt.Sprintf(
				"ig_user(%s){media.after(%s, 500){nodes{code, id},page_info}}",
				userid, lastid)},
			"ref": []string{"users::show"},
		}
		req, err := http.NewRequest("POST", "https://www.instagram.com/query/",
			bytes.NewBufferString(ps.Encode()))
		if err != nil {
			log.Fatal(err)
		}
		{
			u, err := url.Parse("https://www.instagram.com/")
			if err != nil {
				log.Fatal(err)
			}
			var ct string
			for _, c := range cli.Jar.Cookies(u) {
				if c.Name == "csrftoken" {
					ct = c.Value
				}
			}
			if ct == "" {
				log.Fatal("csrftoken not in cookies")
			}
			req.Header.Set("x-csrftoken", ct)
		}
		req.Header.Set("x-instagram-ajax", "1")
		req.Header.Set("x-requested-with", "XMLHttpRequest")
		req.Header.Set("origin", "https://www.instagram.com")
		req.Header.Set("accept-encoding", "gzip, deflate")
		req.Header.Set("accept-language", "en-US,en;q=0.8")
		req.Header.Set("user-agent",
			"Mozilla/5.0 (X11; Linux i686) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/46.0.2490.86 Safari/537.36")
		req.Header.Set("origin", "https://www.instagram.com")
		req.Header.Set("content-type",
			"application/x-www-form-urlencoded; charset=UTF-8")
		req.Header.Set("accept", "application/json, text/javascript, */*; q=0.01")
		req.Header.Set("referer", fmt.Sprintf(
			"https://www.instagram.com/%s/", _user))
		req.Header.Set("authority", "www.instagram.com")
		r, err := cli.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		var data struct {
			Media struct {
				PageInfo struct {
					HasNextPage bool `json:"has_next_page"`
				} `json:"page_info"`
				Nodes []struct {
					Code string `json:"code"`
					ID   string `json:"id"`
				} `json:"nodes"`
			} `json:"media"`
		}
		gr, err := gzip.NewReader(r.Body)
		if err != nil {
			r.Body.Close()
			log.Fatal(err)
		}
		if err := json.NewDecoder(gr).Decode(&data); err != nil {
			r.Body.Close()
			log.Fatal(err)
		}
		r.Body.Close()
		media := data.Media
		for _, m := range media.Nodes {
			getMedia(&cli, fmt.Sprintf(
				"https://www.instagram.com/p/%s/",
				m.Code))
			lastid = m.ID
		}
		hasNext = media.PageInfo.HasNextPage
	}
}
