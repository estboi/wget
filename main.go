package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"wget/app"
)

func main() {
	// Command-line flags
	background := flag.Bool("B", false, "Download in the background")
	output := flag.String("O", "", "Save as a different name")
	downloadDir := flag.String("P", ".", "Download directory")
	rateLimit := flag.String("rate-limit", "", "Download speed limit (e.g., 400k, 2M)")
	inputFile := flag.String("i", "", "File containing multiple download links")
	mirror := flag.Bool("mirror", false, "Mirror a website")
	reject := flag.String("reject", "", "Reject specific file extensions")
	exclude := flag.String("X", "", "Exclude specific directory paths")

	flag.Parse()
	*downloadDir = strings.Replace(*downloadDir, "~", os.Getenv("HOME"), 1)

	url := os.Args[len(os.Args)-1]
	if flag.NArg() > 0 && *inputFile == "" && !*mirror {
		if *background {
			fmt.Println("Output will be written to 'wget-log'")
			app.DownloadInBackground(url)
		} else {
			app.DownloadFile(url, *output, *downloadDir, *rateLimit)
		}
	} else if *inputFile != "" {
		app.DownloadMultipleFiles(*inputFile)
	} else if *mirror {
		app.MirrorWebsite(url, *downloadDir, *reject, *exclude)
	} else {
		fmt.Println("Invalid usage. Use -url, -i, or --mirror flag.")
	}
}
