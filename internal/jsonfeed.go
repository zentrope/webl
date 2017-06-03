// Copyright 2017 Keith Irwin. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type ItemAuthor struct {
	Name string `json:"name"`
}

type FeedItem struct {
	Id            string     `json:"id"`
	Url           string     `json:"url"`
	Title         string     `json:"title"`
	DatePublished string     `json:"date_published"`
	DateModified  string     `json:"date_modified,omitempty"`
	Author        ItemAuthor `json:"author"`
	ContentHtml   string     `json:"content_html"`
	ContentText   string     `json:"content_text,omitempty"`
}

type JSONFeed struct {
	Version     string     `json:"version"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	HomePageURL string     `json:"home_page_url"`
	FeedURL     string     `json:"feed_url"`
	Author      ItemAuthor `json:"author"`
	Items       []FeedItem `json:"items"`
}

func NewJSONFeed(config WebConfig, posts []*LatestPost) (string, error) {

	items := make([]FeedItem, 0)
	for _, p := range posts {
		items = append(items, FeedItem{
			Id:            config.BaseURL + "/post/" + p.UUID,
			Url:           config.BaseURL + "/post/" + p.UUID,
			Title:         p.Slugline,
			DatePublished: p.DateCreated.Format(time.RFC3339),
			DateModified:  p.DateUpdated.Format(time.RFC3339),
			Author:        ItemAuthor{p.Author},
			ContentHtml:   MarkdownToHtml(p.Text),
		})
	}

	feed := JSONFeed{
		Version:     "https://jsonfeed.org/version/1",
		Title:       config.Title,
		Description: fmt.Sprintf("Most recent 40 bloops for '%v'.", config.Title),
		HomePageURL: config.BaseURL,
		FeedURL:     config.BaseURL + "/feeds/json",
		Author:      ItemAuthor{"Root"},
		Items:       items,
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")

	if err := enc.Encode(feed); err != nil {
		return "", err
	}

	return buf.String(), nil
}