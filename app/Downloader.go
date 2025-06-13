package app

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"
)

func DownloadFile(url, output, downloadDir, rateLimit string) {
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
		return
	}

	// Get the file name from the URL
	fileName := output
	if fileName == "" {
		fileName = path.Base(url)
	}
	// Create the output path with the specified download directory
	outputPath := filepath.Join(downloadDir, fileName)

	// Create the output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	defer outFile.Close()

	var writer io.Writer = outFile
	if rateLimit != "" {
		writer = NewRateLimitedWriter(outFile, rateLimit)
	}

	// Get content length
	contentLength := resp.ContentLength
	fmt.Printf("start at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("sending request, awaiting response... status 200 OK\n")
	fmt.Printf("content size: %d [~%.2fMB]\n", contentLength, float64(contentLength)/(1024*1024))
	fmt.Printf("saving file to: %s\n", outputPath)

	buffer := make([]byte, 1024)

	var downloaded int64
	var lastProgressUpdate time.Time

	for {
		n, err := resp.Body.Read(buffer)
		if err != nil && err != io.EOF {
			fmt.Printf("Error downloading: %v\n", err)
			return
		}

		if n == 0 {
			break
		}

		_, err = writer.Write(buffer[:n])
		if err != nil {
			fmt.Printf("Error writing to file: %v\n", err)
			return
		}

		downloaded += int64(n)

		// Update progress based on bytes written to the output file
		now := time.Now()
		if now.Sub(lastProgressUpdate) >= time.Second || downloaded == contentLength {
			lastProgressUpdate = now
			printProgress(downloaded, contentLength)
		}
	}

	fmt.Printf("\nDownloaded [%s]\n", url)
	fmt.Printf("finished at %s\n", time.Now().Format("2006-01-02 15:04:05"))
}

func printProgress(downloaded, total int64) {
	percent := float64(downloaded) / float64(total) * 100
	downloadedMB := float64(downloaded) / (1024 * 1024)
	totalMB := float64(total) / (1024 * 1024)

	barWidth := 40
	completed := int(float64(barWidth) * (percent / 100))
	for i := 0; i < barWidth; i++ {
		if i < completed {
			fmt.Print("=")
		} else {
			fmt.Print(" ")
		}
	}
	fmt.Print("]")
	fmt.Printf("\r%.2f%% [%.2fMB / %.2fMB] [", percent, downloadedMB, totalMB)
}

func DownloadInBackground(url string) {
	logFile, err := os.Create("wget-log")
	if err != nil {
		fmt.Printf("Error creating log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Redirect stdout and stderr to the log file
	os.Stdout = logFile
	os.Stderr = logFile
	DownloadFile(url, "", ".", "")
}

func DownloadMultipleFiles(inputFile string) {
	fmt.Printf("Downloading multiple files from %s\n", inputFile)
	urls, err := readURLsFromFile(inputFile)
	if err != nil {
		fmt.Printf("Error reading URLs from file: %v\n", err)
		return
	}

	// Download each URL in the list
	for _, url := range urls {
		// You can add logic here to specify output names, download directories, etc.
		DownloadFile(url, "", ".", "")
	}

	fmt.Println("Download finished")
}

func readURLsFromFile(inputFile string) ([]string, error) {
	// Implement logic to read URLs from the inputFile and return them as a slice
	// Each line in the file should contain a single URL.

	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
