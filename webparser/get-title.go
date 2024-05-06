package webparser

import (
	"golang.org/x/net/html"
	"log"
	"net/http"
)

const errMsg = "can't load the page title"

func GetTitle(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return errMsg
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("error while closing the response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return errMsg
	}

	tokenizer := html.NewTokenizer(resp.Body)
	for {
		tt := tokenizer.Next()
		switch tt {
		case html.ErrorToken:
			return errMsg
		case html.StartTagToken:
			token := tokenizer.Token()
			if token.Data == "title" {
				if tokenizer.Next() == html.TextToken {
					titleToken := tokenizer.Token()
					return titleToken.Data
				} else {
					return errMsg
				}
			}
		}
		if tt == html.EndTagToken && tokenizer.Token().Data == "html" {
			return errMsg
		}
	}
}
