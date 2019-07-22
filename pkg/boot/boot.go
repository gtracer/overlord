package boot

import (
	"io"
	"log"
	"net/http"
	"os"
)

const k3DownloadUrl = "https://raw.githubusercontent.com/rancher/k3s/master/install.sh"

func Boot() {
	err := fetchK3("k3.sh", k3DownloadUrl)
	if err != nil {
		log.Print("unable to fetch k3 installation script")
	}
}

func fetchK3(filename, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("unable to fetch k3 installation script, err %v", err)
		return err
	}
	defer resp.Body.Close()

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
