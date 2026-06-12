// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever // import "miniflux.app/v2/internal/fever"

import (
	"encoding/xml"
	"net/http"
	"time"

	"miniflux.app/v2/internal/http/response"
)

type baseResponse struct {
	Version       int   `json:"api_version" xml:"api_version"`
	Authenticated int   `json:"auth" xml:"auth"`
	LastRefresh   int64 `json:"last_refreshed_on_time" xml:"last_refreshed_on_time"`
}

func (b *baseResponse) SetCommonValues() {
	b.Version = 3
	b.Authenticated = 1
	b.LastRefresh = time.Now().Unix()
}

/*
The default response is a JSON object containing two members:

	api_version contains the version of the API responding (positive integer)
	auth whether the request was successfully authenticated (boolean integer)

The API can also return XML by passing xml as the optional value of the api argument like so:

http://yourdomain.com/fever/?api=xml

The top level XML element is named response.

The response to each successfully authenticated request will have auth set to 1 and include
at least one additional member:

	last_refreshed_on_time contains the time of the most recently refreshed (not updated)
	feed (Unix timestamp/integer)
*/
func newBaseResponse() baseResponse {
	r := baseResponse{}
	r.SetCommonValues()
	return r
}

func newAuthFailureResponse() baseResponse {
	return baseResponse{Version: 3, Authenticated: 0}
}

type groupsResponse struct {
	baseResponse
	Groups      []group       `json:"groups" xml:"groups>group"`
	FeedsGroups []feedsGroups `json:"feeds_groups" xml:"feeds_groups>feeds_group"`
}

type feedsResponse struct {
	baseResponse
	Feeds       []feed        `json:"feeds" xml:"feeds>feed"`
	FeedsGroups []feedsGroups `json:"feeds_groups" xml:"feeds_groups>feeds_group"`
}

type faviconsResponse struct {
	baseResponse
	Favicons []favicon `json:"favicons" xml:"favicons>favicon"`
}

type itemsResponse struct {
	baseResponse
	Items []item `json:"items" xml:"items>item"`
	Total int    `json:"total_items" xml:"total_items"`
}

type unreadResponse struct {
	baseResponse
	ItemIDs string `json:"unread_item_ids" xml:"unread_item_ids"`
}

type savedResponse struct {
	baseResponse
	ItemIDs string `json:"saved_item_ids" xml:"saved_item_ids"`
}

type group struct {
	ID    int64  `json:"id" xml:"id"`
	Title string `json:"title" xml:"title"`
}

type feedsGroups struct {
	GroupID int64  `json:"group_id" xml:"group_id"`
	FeedIDs string `json:"feed_ids" xml:"feed_ids"`
}

type feed struct {
	ID          int64  `json:"id" xml:"id"`
	FaviconID   int64  `json:"favicon_id" xml:"favicon_id"`
	Title       string `json:"title" xml:"title"`
	URL         string `json:"url" xml:"url"`
	SiteURL     string `json:"site_url" xml:"site_url"`
	IsSpark     int    `json:"is_spark" xml:"is_spark"`
	LastUpdated int64  `json:"last_updated_on_time" xml:"last_updated_on_time"`
}

type item struct {
	ID        int64  `json:"id" xml:"id"`
	FeedID    int64  `json:"feed_id" xml:"feed_id"`
	Title     string `json:"title" xml:"title"`
	Author    string `json:"author" xml:"author"`
	HTML      string `json:"html" xml:"html"`
	URL       string `json:"url" xml:"url"`
	IsSaved   int    `json:"is_saved" xml:"is_saved"`
	IsRead    int    `json:"is_read" xml:"is_read"`
	CreatedAt int64  `json:"created_on_time" xml:"created_on_time"`
}

type favicon struct {
	ID   int64  `json:"id" xml:"id"`
	Data string `json:"data" xml:"data"`
}

// xmlResponse wraps any Fever response struct so that the XML root element
// is always <response>, matching the Fever API specification.
type xmlResponse struct {
	XMLName xml.Name `xml:"response"`
	Any     any      `xml:",any"`
}

// respond sends either a JSON or XML response depending on the "api" query/form
// parameter. When api=xml the body is marshalled as XML with a <response> root
// element; otherwise JSON is returned (the default).
func respond(w http.ResponseWriter, r *http.Request, body any) {
	if r.FormValue("api") == "xml" {
		xmlBytes, err := xml.Marshal(xmlResponse{Any: body})
		if err != nil {
			response.JSONServerError(w, r, err)
			return
		}
		response.XML(w, r, xml.Header+string(xmlBytes))
		return
	}
	response.JSON(w, r, body)
}
