package cmd

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/neoblackcap/ff/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
)

var fetchPutMegaAlbumCmd = &cobra.Command{
	Use:   "putmega",
	Short: "download putmega album",
	Long:  "download putmega album",
	Run:   fetchPutMegaAlbum,
}

type PutMegaParameter struct {
	AlbumUrl    string
	Destination string
	Timeout     int
	ClientID    string
}

var putMegaPara = &PutMegaParameter{}

func init() {
	fetchPutMegaAlbumCmd.Flags().StringVarP(&putMegaPara.AlbumUrl, "album", "a", "", "album url")
	fetchPutMegaAlbumCmd.Flags().IntVarP(&putMegaPara.Timeout, "timeout", "t", 60, "timeout")
	fetchPutMegaAlbumCmd.Flags().StringVarP(&putMegaPara.Destination, "destination", "d", "", "destination")

	err := fetchPutMegaAlbumCmd.MarkFlagRequired("album")
	if err != nil {
		log.Fatalf("init err: %s", err)
	}

	err = fetchPutMegaAlbumCmd.MarkFlagRequired("destination")
	if err != nil {
		log.Fatalf("init err: %s", err)
	}

	rootCmd.AddCommand(fetchPutMegaAlbumCmd)
}

func fetchPutMegaAlbum(cmd *cobra.Command, args []string) {
	newUrl := putMegaPara.AlbumUrl
	log.Info("putmega album url: ", newUrl)
	method := "GET"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, newUrl, payload)

	if err != nil {
		log.Fatal(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	imgUrls := make([]string, 0)
	doc.Find("#content-listing-tabs .list-item-image.fixed-size a.image-container img").Each(func(i int, selection *goquery.Selection) {
		attr, exists := selection.Attr("src")
		if exists {
			log.Infof("img %d url: %s", i, attr)
			u, err2 := url.Parse(attr)
			if err2 != nil {
				log.Fatal("parse image url error ", err2)
			}

			paths := strings.Split(u.Path, "/")
			lastUri := paths[len(paths)-1]
			paths = paths[0:len(paths)-1]
			idents := strings.Split(lastUri, ".")
			lastExt := idents[len(idents) -1]
			newLastUri := fmt.Sprintf("%s.%s", idents[0], lastExt)
			paths = append(paths, newLastUri)
			u.Path = path.Join(paths...)

			log.Info("try to download ", u.String())
			imgUrls = append(imgUrls, u.String())
		}
	})

	nowStr := time.Now().Format("2006-01-02")
	dest := path.Join(putMegaPara.Destination, nowStr)
	folderPath := utils.CreateFolder(dest)
	log.Infof("path %s", folderPath)

	ctx := context.Background()
	wg := &sync.WaitGroup{}
	wg.Add(len(imgUrls))
	for _, imgUrl := range imgUrls {
		go utils.DownloadImage(ctx, imgUrl, folderPath, wg)
	}
	wg.Wait()

}
