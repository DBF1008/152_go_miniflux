// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json // import "miniflux.app/v2/internal/reader/json"

import "testing"

func TestIsJSONFeedV1(t *testing.T) {
	data := []byte(`{
		"version": "https://jsonfeed.org/version/1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": []
	}`)
	if !IsJSONFeed(data) {
		t.Error("Expected IsJSONFeed to return true for a valid JSON Feed v1")
	}
}

func TestIsJSONFeedV11(t *testing.T) {
	data := []byte(`{
		"version": "https://jsonfeed.org/version/1.1",
		"title": "My Example Feed",
		"home_page_url": "https://example.org/",
		"feed_url": "https://example.org/feed.json",
		"items": []
	}`)
	if !IsJSONFeed(data) {
		t.Error("Expected IsJSONFeed to return true for a valid JSON Feed v1.1")
	}
}

func TestIsJSONFeedRejectsPlainJSON(t *testing.T) {
	data := []byte(`{
		"name": "Some API Response",
		"status": "ok",
		"results": [{"id": 1, "value": "hello"}]
	}`)
	if IsJSONFeed(data) {
		t.Error("Expected IsJSONFeed to return false for plain JSON without version field")
	}
}

func TestIsJSONFeedRejectsUnrelatedVersion(t *testing.T) {
	data := []byte(`{
		"version": "2.0",
		"name": "Some API Response"
	}`)
	if IsJSONFeed(data) {
		t.Error("Expected IsJSONFeed to return false when version does not contain 'jsonfeed'")
	}
}

func TestIsJSONFeedRejectsInvalidJSON(t *testing.T) {
	data := []byte(`this is not valid json`)
	if IsJSONFeed(data) {
		t.Error("Expected IsJSONFeed to return false for invalid JSON")
	}
}

func TestIsJSONFeedRejectsEmptyInput(t *testing.T) {
	if IsJSONFeed([]byte(``)) {
		t.Error("Expected IsJSONFeed to return false for empty input")
	}
}

func TestIsJSONFeedRejectsJSONArray(t *testing.T) {
	data := []byte(`[{"version": "https://jsonfeed.org/version/1"}]`)
	if IsJSONFeed(data) {
		t.Error("Expected IsJSONFeed to return false for a JSON array")
	}
}

func TestIsJSONFeedRejectsEmptyObject(t *testing.T) {
	data := []byte(`{}`)
	if IsJSONFeed(data) {
		t.Error("Expected IsJSONFeed to return false for an empty JSON object")
	}
}

func TestIsJSONFeedRejectsNullVersion(t *testing.T) {
	data := []byte(`{"version": null, "title": "Not a feed"}`)
	if IsJSONFeed(data) {
		t.Error("Expected IsJSONFeed to return false when version is null")
	}
}
