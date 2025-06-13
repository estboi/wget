package app

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

func MirrorWebsite(url, downloadDir, reject, exclude string) {
	fmt.Printf("Mirroring website: %s\n", url)

	rejectSuffixes := strings.Split(reject, ",")
	excludePaths := strings.Split(exclude, ",")

	// Create a directory with the domain name as the folder name
	domainName := getDomainName(url)
	siteDir := path.Join(downloadDir, domainName)
	if err := os.MkdirAll(siteDir, os.ModePerm); err != nil {
		fmt.Printf("Error creating site directory: %v\n", err)
		return
	}

	// Ensure that the siteDir exists before proceeding
	if _, err := os.Stat(siteDir); os.IsNotExist(err) {
		fmt.Printf("Error: Site directory %s does not exist\n", siteDir)
		return
	}

	// Perform an HTTP request to fetch the website content
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error fetching website: %v\n", err)
		return
	}
	defer resp.Body.Close()

	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading website content: %v\n", err)
		return
	}
	links, paths := findLinks(string(htmlContent))
	htmlString := string(htmlContent)

	for _, path := range paths {
		if shouldDownload(path, rejectSuffixes, excludePaths) {
			if err := os.MkdirAll(siteDir+"/"+path, os.ModePerm); err != nil {
				fmt.Printf("Error creating site directory: %v\n", err)
				return
			}
		}
	}

	for _, link := range links {
		if shouldDownload(link, rejectSuffixes, excludePaths) {
			fmt.Printf("Downloading file: %s\n", link)
			DownloadFile(url+link, link, siteDir, "")
			htmlString = strings.Replace(htmlString, link, strings.TrimLeft(link, "/"), -1)
			fmt.Println(strings.TrimLeft(link, "/"))
		}
	}

	file, err := os.Create(path.Join(siteDir, "index.html"))
	if err != nil {
		fmt.Printf("Error creating index.html file: %v\n", err)
	}

	_, err = file.Write([]byte(htmlString))
	if err != nil {
		fmt.Printf("Error writing index.html: %v\n", err)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Status %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
		return
	}
}

func findLinks(html string) ([]string, []string) {
	var links []string
	var paths []string

	srcReg := regexp.MustCompile(`(src|href)=['"]([0-9a-zA-Z./_-]+)['"]`)
	urlReg := regexp.MustCompile(`url\(['"]([0-9a-zA-Z./_-]+)['"]\)`)

	hrefMatches := srcReg.FindAllStringSubmatch(html, -1)
	for _, match := range hrefMatches {
		if len(match) == 3 {
			// Match found, append the link
			links = append(links, match[2])

			// Extract the folder path (if any) and append it to paths
			if parts := strings.Split(match[2], "/"); len(parts) > 1 {
				paths = append(paths, strings.Join(parts[:len(parts)-1], "/"))
			}
		}
	}

	imgSrcMatches := urlReg.FindAllStringSubmatch(html, -1)
	for _, match := range imgSrcMatches {
		if len(match) == 2 {
			// Match found, append the link
			links = append(links, match[1])

			// Extract the folder path (if any) and append it to paths
			if parts := strings.Split(match[1], "/"); len(parts) > 1 {
				paths = append(paths, strings.Join(parts[:len(parts)-1], "/"))
			}
		}
	}

	return links, paths
}

func getDomainName(urlString string) string {
	pattern := `^(?:https?://)?(www\.)?([a-zA-Z0-9.-]+)`
	regex := regexp.MustCompile(pattern)

	// Find the first submatch that corresponds to the domain name
	matches := regex.FindStringSubmatch(urlString)
	if len(matches) >= 3 {
		return matches[2]
	}

	fmt.Printf("Error extracting domain name from URL: %s\n", urlString)
	return ""
}

func shouldDownload(link string, rejectSuffixes []string, excludePaths []string) bool {
	// Check if the link should be rejected based on file suffixes
	for _, suffix := range rejectSuffixes {
		if suffix != "" {
			if strings.HasSuffix(link, suffix) {
				fmt.Printf("Rejected due to suffix: %s\n", link)
				return false
			}
		}
	}

	// Check if the link should be excluded based on paths
	for _, path := range excludePaths {
		if path != "" {
			if strings.HasPrefix(link, "."+path) {
				fmt.Printf("Excluded path: %s\n", link)
				return false
			}
		}
	}

	return true
}
