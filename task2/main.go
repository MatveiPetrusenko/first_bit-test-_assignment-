package main

import (
	"encoding/json"
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Config struct {
	ConcurrentDownloads int      `json:"concurrent_downloads"` //the maximum number of interchangeable files
	DownloadAttempts    int      `json:"download_attempts"`    //number of attempts to download each file, in case errors
	Urls                []string `json:"urls"`
}

func main() {
	configFile, downloadDir := checkArguments()

	if !checkDir(downloadDir) {
		log.Fatalln("Download directory does not exist")
	}

	config, err := readConfig(configFile)
	if err != nil {
		log.Fatalln("Error reading config:", err)
	}

	var wg sync.WaitGroup
	downloadSemaphore := make(chan struct{}, config.ConcurrentDownloads) //make channel with capacity equal config.ConcurrentDownloads

	for _, url := range config.Urls {
		wg.Add(1)

		go func(url string) {
			defer wg.Done() // reduce count of waitGroup after go func complete or something happen
			var attempts int

			for {
				if attempts >= config.DownloadAttempts {
					fmt.Printf("Error downloading %s: maximum download attempts exceeded\n", url)
					return
				}

				err := downloadFile(url, downloadDir, downloadSemaphore)
				if err != nil {
					fmt.Printf("Error downloading %s: %s\n", url, err)
					attempts++
					continue
				}
				break
			}
		}(url)
	}
	wg.Wait()
}

// checkArguments checks if we have received 3 arguments and if yes, return configFile + downloadDir as string
func checkArguments() (string, string) {
	if len(os.Args) < 3 {
		log.Fatalln("Usage: go run main.go config.json /path/to/download/directory")
		os.Exit(1)
	}

	configFile := os.Args[1]
	downloadDir := os.Args[2]

	return configFile, downloadDir
}

// checkDir checks directory route exists or does not exist
func checkDir(inputDir string) bool {
	_, err := os.Stat(inputDir)
	if os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// readConfig reading json from file and return unmarshalling json as struct
func readConfig(configFile string) (Config, error) {
	var config Config

	file, err := os.Open(configFile)
	if err != nil {
		return config, err
	}
	defer file.Close()

	decodeJson := json.NewDecoder(file)
	err = decodeJson.Decode(&config)
	if err != nil {
		return config, err
	}

	// if url is escaping with such characters
	for k, _ := range config.Urls {
		config.Urls[k] = strings.Trim(config.Urls[k], "<>")
	}

	return config, nil
}

// downloadFile connects by url and creates a download file with a download indicator
func downloadFile(url string, downloadDir string, downloadSemaphore chan struct{}) error {
	downloadSemaphore <- struct{}{}
	defer func() {
		<-downloadSemaphore
	}()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	filename := filepath.Base(url) // get name from url
	filepath := filepath.Join(downloadDir, filename)
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Create progress bar for downloading file
	progressBar := progressbar.DefaultBytes(
		resp.ContentLength,
		fmt.Sprintf("Downloading %s: ", filename),
	)

	// Create multi writer to write to both file and progress bar
	writer := io.MultiWriter(out, progressBar)

	// Write the body to file and progress bar
	_, err = io.Copy(writer, resp.Body)
	if err != nil {
		os.Remove(filepath) // delete file if something happen
		return err
	}

	fmt.Printf("\nDownloaded %s\n", filename)
	return nil
}
