package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/neoblackcap/ff/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net/http"
	urlm "net/url"
	"strings"
	"sync"
	"time"
)

var fetchImgCreditAlbumCmd = &cobra.Command{
	Use:   "imgcredit",
	Short: "download imgcredit.xyz images",
	Run:   fetchImgCreditAlbum,
}

var albumUrl string
var destination string
var timeout int

func init() {
	fetchImgCreditAlbumCmd.Flags().StringVarP(&albumUrl, "album", "a", "", "album url")
	fetchImgCreditAlbumCmd.Flags().IntVarP(&timeout, "timeout", "t", 60, "timeout")
	fetchImgCreditAlbumCmd.Flags().StringVarP(&destination, "destination", "d", "", "destination")

	err := fetchImgCreditAlbumCmd.MarkFlagRequired("album")
	if err != nil {
		log.Fatalf("init err: %s", err)
	}

	err = fetchImgCreditAlbumCmd.MarkFlagRequired("destination")
	if err != nil {
		log.Fatalf("init err: %s", err)
	}

	rootCmd.AddCommand(fetchImgCreditAlbumCmd)
}

func fetchImgCreditAlbum(cmd *cobra.Command, args []string) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancelFunc()
	url := strings.TrimSuffix(albumUrl, "/") + "/?sort=title_asc&page=1"
	imageUrls := fetchAndParseAlbum(ctx, url)
	var wg sync.WaitGroup

	var dest string
	s := len(imageUrls)
	if s > 0 {
		dest = utils.CreateFolder(destination)
		wg.Add(s)
	} else {
		log.Info("no image found")
	}

	for _, u := range imageUrls {
		go fetchImage(ctx, u, dest, &wg)
	}

	if s > 0 {
		wg.Wait()
	} else {
		log.Info("do nothing")
	}

}

func fetchAndParseAlbum(ctx context.Context, url string) []string {
	client := http.Client{}

	// create request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatal("create request error: ", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	// send request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("get album error: ", err)
	}

	// parse
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("album parse error: ", err)
	}
	images := doc.Find(".image-container")

	images1 := make([]string, 0)
	images.Each(func(i int, selection *goquery.Selection) {
		link, exists := selection.Attr("href")
		if exists {
			images1 = append(images1, link)
		}
	})

	nextUrl, err := getNextPage(doc)

	var images2 []string
	switch err {
	case ErrNoNextPage:
	case nil:
		images2 = fetchAndParseAlbum(ctx, nextUrl)
	default:
		log.Fatal("fetch next page url error: ", err)
	}

	if len(images2) > 0 {
		images1 = append(images1, images2...)
	}

	return images1

}

var ErrNoNextPage = errors.New("no next page")

func getNextPage(doc *goquery.Document) (string, error) {
	s := doc.Find(".pagination-next > a")
	if s.Size() != 0 {
		link, exists := s.Attr("href")
		if exists {
			return link, nil
		} else {
			return "", ErrNoNextPage
		}

	} else {
		return "", ErrNoNextPage
	}
}

func fetchImage(ctx context.Context, url string, dest string, wg *sync.WaitGroup) {
	defer wg.Done()
	client := http.Client{}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		log.Fatal("create fetch image request error: ", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("fetch image error: ", err)
	}

	root, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("parse single image")
		return
	}
	selection := root.Find(".btn.btn-download.default")
	if selection.Size() != 1 {
		log.Fatal("multiple download button found")
		return
	}

	link, exists := selection.Attr("href")
	if exists {
		r, err := http.NewRequestWithContext(ctx, "GET", link, nil)
		if err != nil {
			log.Fatal("create fetch image request error: ", err)
		}
		r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:79.0) Gecko/20100101 Firefox/79.0")
		r.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		rp, err := client.Do(r)
		if err != nil {
			log.Fatal("fetch image error: ", err)
		}
		parse, err := urlm.Parse(link)
		if err != nil {
			log.Fatal("parse image download link error: ", err)
		}

		uris := strings.Split(parse.Path, "/")
		filename := strings.Split(parse.Path, "/")[len(uris)-1]
		body, err := ioutil.ReadAll(rp.Body)
		output := fmt.Sprintf("%s/%s", dest, filename)

		log.Infof("write file %s", output)
		_ = ioutil.WriteFile(output, body, 0644)

	}

}
