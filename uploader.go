package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
)

type YandexUpload struct {
	Operation_id string `json:"operation_id"`
	Href         string `json:"href"`
	Method       string `json:"method"`
	Templated    bool   `json:"templated"`
}
type ErrorResponse struct {
	Message     string `json:"message"`
	Description string `json:"description"`
	Error       string `json:"error"`
}

type ByModTime []os.FileInfo

const YandexOauth = "{TOKEN}}"
const YandexUploadRequestUrl = "https://cloud-api.yandex.net/v1/disk/resources/upload?path={UPLOADDIR}"
const BackupPath = "{FILEPATH}"

func main() {
	req, err := http.NewRequest("GET", YandexUploadRequestUrl+latestBackup(), nil)
	req.Header.Add("Authorization", "OAuth "+YandexOauth)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		var t YandexUpload
		json.Unmarshal(body, &t)
		data, err := os.Open(BackupPath + latestBackup())
		if err != nil {
			fmt.Println(err)
		}
		defer data.Close()
		req, err := http.NewRequest(t.Method, t.Href, data)
		resp, err := client.Do(req)
		if resp.StatusCode == 201 {
			fmt.Println("Success")
		}
		if resp.StatusCode != 201 {
			fmt.Println("Failed")
			//TODO: OS Event log
		}
	}
	if resp.StatusCode == 409 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		var t ErrorResponse
		json.Unmarshal(body, &t)
		fmt.Printf("Error: %#v\r\n", t)
	}
}

func (fis ByModTime) Len() int {
	return len(fis)
}

func (fis ByModTime) Swap(i, j int) {
	fis[i], fis[j] = fis[j], fis[i]
}

func (fis ByModTime) Less(i, j int) bool {
	return fis[i].ModTime().Before(fis[j].ModTime())
}

func latestBackup() string {
	f, _ := os.Open(BackupPath)
	fis, _ := f.Readdir(-1)
	f.Close()
	sort.Sort(ByModTime(fis))
	return fis[len(fis)-1].Name()
}
