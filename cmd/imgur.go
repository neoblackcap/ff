package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/neoblackcap/ff/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

var fetchImgurAlbumCmd = &cobra.Command{
	Use:   "imgur",
	Short: "download imgur album",
	Long:  "download imgur album",
	Run:   fetchImageAlbum,
}

type ImgurParameter struct {
	AlbumUrl    string
	Destination string
	Timeout     int
	ClientID string
}

var imgurPara = &ImgurParameter{}

type Result struct {
	Data    Data `json:"data"`
	Status  int  `json:"status"`
	Success bool `json:"success"`
}

type Result2 struct {
	Data    []Image `json:"data"`
	Status  int     `json:"status"`
	Success bool    `json:"success"`
}

type Data struct {
	AccountId       string   `json:"account_id"`
	AccountUrl      string   `json:"account_url"`
	AdConfig        AdConfig `json:"ad_config"`
	Cover           string   `json:"cover"`
	CoverEdited     *bool    `json:"cover_edited"`
	CoverHeight     float64  `json:"cover_height"`
	CoverWidth      float64  `json:"cover_width"`
	Datetime        float64  `json:"datetime"`
	Description     string   `json:"description"`
	Favorite        bool     `json:"favorite"`
	Id              string   `json:"id"`
	Images          []Image  `json:"images"`
	ImageCount      float64  `json:"images_count"`
	InGallery       bool     `json:"in_gallery"`
	IncludeAlbumAds bool     `json:"include_album_ads"`
	IsAd            bool     `json:"is_ad"`
	IsAlbum         bool     `json:"is_album"`
	Layout          string   `json:"layout"`
	Link            string   `json:"link"`
	Privacy         string   `json:"privacy"`
	Section         string   `json:"section"`
	Title           string   `json:"title"`
	Views           float64  `json:"views"`
}

type AdConfig struct {
	HighRiskFlags   []string `json:"highRiskFlags"`
	SafeFlags       []string `json:"safeFlags"`
	ShowsAds        bool     `json:"showsAds"`
	UnsafeFlags     []string `json:"unsafeFlags"`
	WallUnsafeFlags []string `json:"wallUnsafeFlags"`
}

type Image struct {
	AccountId   string   `json:"account_id"`
	AccountUrl  string   `json:"account_url"`
	AdType      float64  `json:"ad_type"`
	AdUrl       string   `json:"ad_url"`
	Animated    bool     `json:"animated"`
	Bandwidth   float64  `json:"bandwidth"`
	Datetime    float64  `json:"datetime"`
	Description string   `json:"description"`
	Edited      string   `json:"edited"`
	Favorite    bool     `json:"favorite"`
	HasSound    bool     `json:"has_sound"`
	Height      float64  `json:"height"`
	Id          string   `json:"id"`
	InGallery   bool     `json:"in_gallery"`
	InMostViral bool     `json:"in_most_viral"`
	IsAd        bool     `json:"is_ad"`
	Link        string   `json:"link"`
	Nsfw        string   `json:"nsfw"`
	Section     string   `json:"section"`
	Size        float64  `json:"size"`
	Tags        []string `json:"tags"`
	Title       string   `json:"title"`
	Type        string   `json:"type"`
	Views       float64  `json:"views"`
	Vote        string   `json:"vote"`
	Width       float64  `json:"width"`
}

func init() {
	fetchImgurAlbumCmd.Flags().StringVarP(&imgurPara.AlbumUrl, "album", "a", "", "album url")
	fetchImgurAlbumCmd.Flags().IntVarP(&imgurPara.Timeout, "timeout", "t", 60, "timeout")
	fetchImgurAlbumCmd.Flags().StringVarP(&imgurPara.Destination, "destination", "d", "", "destination")
	fetchImgurAlbumCmd.Flags().StringVarP(&imgurPara.ClientID, "client-id", "i", "", "imgur developer client id")

	err := fetchImgurAlbumCmd.MarkFlagRequired("album")
	if err != nil {
		log.Fatalf("init err: %s", err)
	}

	err = fetchImgurAlbumCmd.MarkFlagRequired("destination")
	if err != nil {
		log.Fatalf("init err: %s", err)
	}

	err = fetchImgurAlbumCmd.MarkFlagRequired("client-id")
	if err != nil {
		log.Fatalf("init err: %s", err)
	}

	rootCmd.AddCommand(fetchImgurAlbumCmd)
}

func fetchAlbum(albumHash string) ([]byte, error) {

	urlPara := "https://api.imgur.com/3/album/%s/images"
	newUrl := fmt.Sprintf(urlPara, albumHash)
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
		fmt.Println(err)
	}
	authorizationHeader := fmt.Sprintf("Client-ID %s", imgurPara.ClientID)
	req.Header.Add("Authorization", authorizationHeader)

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return body, err
}

func fetchImgurImage(ctx context.Context, image Image, dst string, wg *sync.WaitGroup) {
	defer wg.Done()
	method := "GET"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, image.Link, payload)

	if err != nil {
		fmt.Println(err)
	}

	authorizationHeader := fmt.Sprintf("Client-ID %s", imgurPara.ClientID)
	req.Header.Add("Authorization", authorizationHeader)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	u, err := url.Parse(image.Link)
	if err != nil {
		fmt.Print(image.Link)
		return
	}
	paths := strings.Split(u.Path, "/")
	_filename := paths[len(paths)-1]

	filename := fmt.Sprintf("%s/%s", dst, _filename)

	_ = ioutil.WriteFile(filename, body, 0644)
}

func fetchImageAlbum(cmd *cobra.Command, args []string) {
	// albumHash := "B88tSQQ#"
	parse, err := url.Parse(imgurPara.AlbumUrl)
	if err != nil {
		log.Fatal("parse imgur album url error: ", err)
	}

	paths := strings.Split(parse.Path, "/")

	albumHash := paths[len(paths)-1]
	log.Info("album hash: ", albumHash)

	data, _ := fetchAlbum(albumHash)
	// data := fakeData()
	var result Result
	var result2 Result2
	var images []Image

	// try to unmarshal data with Result
	err = json.Unmarshal(data, &result)
	if err == nil {
		images = result.Data.Images
	}

	if images == nil {
		err = json.Unmarshal(data, &result2)
		if err == nil {
			images = result2.Data
		}
	}

	if images == nil {
		log.Fatalf("parse imgur sdk return data error: %s, return data: %s", err, data)
	}

	dst := utils.CreateFolder(imgurPara.Destination)

	var wg sync.WaitGroup
	wg.Add(len(images))
	ctx, _ := context.WithTimeout(context.Background(), time.Second * time.Duration(imgurPara.Timeout))

	for _, image := range images {
		go fetchImgurImage(ctx, image, dst, &wg)
	}
	wg.Wait()
	log.Info("fetch all images")
}
