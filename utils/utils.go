package utils

import (
	"bytes"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"sync"
)

func CreateFolder(prefix string) string {
	_, err := os.Stat(prefix)
	if os.IsNotExist(err) {
		err := os.MkdirAll(prefix, 0755)
		if err != nil {
			log.Fatal("create folder failure: ", err)
		}
	}
	return prefix
}

func createFolderWithIndex(prefix string, index int) string {
	folder := fmt.Sprintf("%s_%d", prefix, index)
	_, err := os.Stat(folder)
	if os.IsNotExist(err) {
		_ = os.Mkdir(folder, 0744)
		return folder
	}
	return createFolderWithIndex(prefix, index+1)

}

func DownloadImage(ctx context.Context, url string, dst string, wg *sync.WaitGroup) {
	defer wg.Done()
	method := "GET"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, method, url, payload)

	if err != nil {
		fmt.Println(err)
	}

	res, err := client.Do(req)
	if err != nil {
		log.Errorf("fetch image<%s> error: %s", url, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("fetch image<%s> error: %s", url, err)
	}
	paths := strings.Split(url, "/")
	_filename := paths[len(paths)-1]
	names := strings.Split(_filename, ".")
	ext := names[len(names)-1]
	name := names[0]

	filename := fmt.Sprintf("%s/%s.%s", dst, name, ext)

	_ = ioutil.WriteFile(filename, body, 0644)
	log.Infof("image<%s> downloaded", url)
}
