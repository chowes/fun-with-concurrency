package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type crawlResult struct {
	depth int
	links []string
}

type searchURL struct {
	depth int
	url   string
}

// given an HTML page, traverse the tree of elements and return a list of URLs
func findLinks(node *html.Node) []string {
	var links []string

	if node.Type == html.ElementNode && node.Data == "a" {
		for _, a := range node.Attr {
			if a.Key == "href" {
				links = append(links, a.Val)
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		links = append(links, findLinks(child)...)
	}

	return links
}

// given a URL, retrieve the HTML and extract all URLs
func crawl(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get URL %q, status=%d", url, resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	var links []string
	for _, l := range findLinks(doc) {
		link, err := resp.Request.URL.Parse(l)
		if err != nil {
			continue
		}
		links = append(links, link.String())
	}

	return links, nil
}

func main() {
	startPages := flag.String("start-pages", "", "comma separated list of URLs to crawl")
	maxDepth := flag.Int("depth", 0, "max number of links to follow")
	flag.Parse()

	startURLs := strings.Split(*startPages, ",")

	seenLinks := map[string]bool{}
	for _, url := range startURLs {
		seenLinks[url] = true
	}

	linksToSearch := make(chan searchURL)
	foundLinks := make(chan crawlResult)
	var wg sync.WaitGroup

	go func() {
		for _, url := range startURLs {
			linksToSearch <- searchURL{
				url:   url,
				depth: 0,
			}
		}
	}()

	numCrawlers := 20
	for _ = range numCrawlers {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for url := range linksToSearch {
				if url.depth >= *maxDepth {
					return
				}

				links, err := crawl(url.url)
				if err != nil {
					log.Printf("failed to crawl %q: %v", url, err)
				}

				foundLinks <- crawlResult{
					links: links,
					depth: url.depth,
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(foundLinks)
	}()

	for result := range foundLinks {
		for _, l := range result.links {
			if seenLinks[l] {
				continue
			}
			seenLinks[l] = true
			fmt.Printf("Found: %q\n", l)

			go func() {
				linksToSearch <- searchURL{
					depth: result.depth + 1,
					url:   l,
				}
			}()
		}
	}
}
