package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	readability "github.com/go-shiori/go-readability"
)

const version = "v0_20221002"
const tmpDir = "link2epubtmpdir"

func main() {
	versionFlag := flag.Bool("v", false, "version")
	linkFlag := flag.String("l", "", "Link to download")
	typeFlag := flag.String("type", "mobi", "Type of epub. Available: mobi (default), epub")
	titleFlag := flag.String("title", "", "Title is automatically getted from article, if want to change it, use this flag")

	flag.Parse()

	fmt.Println("link2epub version:", version)
	if *versionFlag {
		os.Exit(0)
	}

	if *typeFlag != "mobi" && *typeFlag != "epub" {
		log.Fatal("not valid type")
	}
	err := os.Mkdir(tmpDir, os.ModePerm)
	if err != nil {
		log.Printf("error creating tmp dir %s: %v\nRemoving it and continuing.", tmpDir, err)
		err = os.RemoveAll(tmpDir)
		if err != nil {
			log.Fatalf("err removing %s: %v\n", tmpDir, err)
		}
	}

	// get link
	fmt.Println("\n> getting the link")
	resp, err := http.Get(*linkFlag)
	if err != nil {
		log.Fatalf("failed to download %s: %v\n", *linkFlag, err)
	}
	defer resp.Body.Close()

	// convert the html to simple html with go-readability
	article, err := readability.FromReader(resp.Body, *linkFlag)
	if err != nil {
		log.Fatalf("failed to parse %s: %v\n", *linkFlag, err)
	}

	fmt.Printf("	URL     : %s\n", *linkFlag)
	fmt.Printf("	Title   : %s\n", article.Title)
	fmt.Printf("	Author  : %s\n", article.Byline)
	fmt.Printf("	Length  : %d\n", article.Length)
	fmt.Printf("	Excerpt : %s\n", article.Excerpt)
	fmt.Printf("	SiteName: %s\n", article.SiteName)
	fmt.Printf("	Image   : %s\n", article.Image)
	fmt.Printf("	Favicon : %s\n", article.Favicon)

	// get images
	fmt.Println("\n>getting the images")
	imgRegex := regexp.MustCompile(`(<img )([^>]*)(src=")([^"]*)"`)
	imgs := imgRegex.FindAllSubmatch([]byte(article.Content), -1)
	for i, img := range imgs {
		fmt.Println("	img", i, string(img[4]))
		filename, err := downloadImg(string(img[4]), strconv.Itoa(i))
		if err != nil {
			log.Fatalf("error in downloadImg %s: %v\n", img[4], err)
		}

		// replace in the article.Content the current img by new filename
		article.Content = strings.Replace(article.Content, string(img[4]), filename, -1)
	}

	if *titleFlag != "" {
		article.Title = *titleFlag
	}

	// add title to content
	article.Content = `
		<h1>` + article.Title + `</h1>
		<h2 style="text-align:right;">` + article.Byline + `</h2>
		<br>
	` + article.Content

	// store html file
	filename := article.Title + " - " + article.Byline
	out, err := os.Create(tmpDir + "/" + filename + ".html")
	if err != nil {
		log.Fatalf("failed creating index.xhtml: %v\n", err)
	}
	defer out.Close()

	_, err = out.Write([]byte(article.Content))
	if err != nil {
		log.Fatalf("failed writting index.html: %v\n", err)
	}
	out.Sync()

	// call calibre to convert the html to epub/mobi
	fmt.Println("\n>converting to", *typeFlag)
	cmd := exec.Command("ebook-convert", tmpDir+"/"+filename+".html", filename+"."+*typeFlag)
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed converting the html to %s: %v\n", *typeFlag, err)
	}

	// delete tmp dir
	cmd = exec.Command("rm", "-rf", tmpDir)
	if err := cmd.Run(); err != nil {
		log.Fatalf("failed removing the tmp dir %s: %v\n", tmpDir, err)
	}
}

func downloadImg(url string, path string) (string, error) {
	url = strings.Replace(url, "/max/60/", "/max/1000/", -1) // for "medium.com" api
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(body)
	filename := path + "." + strings.Replace(contentType, "image/", "", -1)

	out, err := os.Create(tmpDir + "/" + filename)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = out.Write(body)
	if err != nil {
		return "", err
	}
	out.Sync()

	return filename, nil
}
