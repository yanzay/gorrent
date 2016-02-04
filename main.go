package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"sync"

	"golang.org/x/net/html/charset"
)

var (
	watchdir = flag.String("watchdir", "", "Watch dir to download torrents")
	link     = flag.String("link", "http://torrentrss.net/getrss.php?rsslink=rPEELP", "Link to RSS")
)

type RSS struct {
	Items []Item `xml:"channel>item"`
}

type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}

func main() {
	flag.Parse()
	if *watchdir == "" {
		fmt.Println("watchdir parameter required")
		return
	}
	var feed RSS
	resp, err := http.Get("http://torrentrss.net/getrss.php?rsslink=rPEELP")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&feed)
	if err != nil {
		log.Fatal(err)
	}
	wg := &sync.WaitGroup{}
	for _, item := range feed.Items {
		wg.Add(1)
		go download(item.Link, wg)
	}
	wg.Wait()
}

func download(link string, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := http.Get(link)
	defer resp.Body.Close()
	if err != nil {
		log.Printf("Can't download torrent file: %s", err.Error())
		return
	}
	disposition := resp.Header.Get("Content-Disposition")
	_, params, err := mime.ParseMediaType(disposition)
	if err != nil {
		log.Printf("Can't get filename from response (%s): %s", link, err)
		return
	}
	filename := *watchdir + "/" + params["filename"]
	_, err = os.Stat(filename)
	if os.IsNotExist(err) {
		log.Printf("Creating file %s", filename)
		file, err := os.Create(filename)
		if err != nil {
			log.Println(err)
			return
		}
		defer file.Close()
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Printf("Error copying file: %s", err)
		}
	}
}
