// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package json // import "miniflux.app/v2/internal/reader/json"

import (
	"encoding/json"
	"strings"
)

// IsJSONFeed reports whether data is a valid JSON Feed.
//
// Per the JSON Feed specification (https://www.jsonfeed.org/version/1.1/),
// a conforming feed MUST include a "version" field whose value is the URL
// of the version of the format it uses (e.g. "https://jsonfeed.org/version/1.1").
//
// This function is intentionally stricter than DetectFeedFormat — which
// accepts any JSON object — so that well-known-URL probing does not
// mistake ordinary JSON endpoints for feed subscriptions.
func IsJSONFeed(data []byte) bool {
	var feed struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(data, &feed); err != nil {
		return false
	}
	return strings.Contains(feed.Version, "jsonfeed")
}
