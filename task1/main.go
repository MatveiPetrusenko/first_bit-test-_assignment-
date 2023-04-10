package main

import (
	"bufio"
	"fmt"
	"github.com/dustin/go-humanize"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	out := bufio.NewWriter(os.Stdout)
	defer out.Flush()

	var inputDir string
	_, err := fmt.Fscan(in, &inputDir)
	if err != nil {
		log.Fatal(err)
	}

	if checkDir(inputDir) {
		err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			switch runtime.GOOS {
			case "windows":
				cmd := exec.Command("powershell", "(Get-Item "+path+").CreationTime.ToString(\"dd.MM.yyyy HH:mm:ss\")")
				output, err := cmd.Output()
				if err != nil {
					fmt.Println(err)
				}
				creationTime := strings.TrimSpace(string(output))

				fmt.Println("\nName:", info.Name(), "\nSize:", humanize.Bytes(uint64(info.Size())), "\nDate:", creationTime)
				fmt.Println(path)
			case "linux":
				cmd := exec.Command("stat", "-c", "%w", path)
				output, err := cmd.Output()
				if err != nil {
					fmt.Println(err)
				}

				creationTime := strings.TrimSpace(string(output))
				layout := "2006-01-02 15:04:05.999999999 -0700"
				t, err := time.Parse(layout, creationTime)
				if err != nil {
					fmt.Println(err)
				}

				fmt.Println("\nName:", info.Name(), "\nSize:", humanize.Bytes(uint64(info.Size())), "\nDate:", t.Format("02.01.2006 15:04:05"))
				fmt.Println(path)
			default:
				fmt.Println("Operating system does not support")
			}

			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatalf("%s\n%s", inputDir, "Directory Not Exist")
	}
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
