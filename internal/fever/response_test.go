// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fever // import "miniflux.app/v2/internal/fever"

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func sampleGroupsResponse() groupsResponse {
	var result groupsResponse
	result.Groups = []group{{ID: 1, Title: "Tech"}}
	result.FeedsGroups = []feedsGroups{{GroupID: 1, FeedIDs: "10,11"}}
	result.SetCommonValues()
	return result
}

func TestWantsXMLResponse(t *testing.T) {
	scenarios := []struct {
		name     string
		method   string
		target   string
		body     string
		expected bool
	}{
		{"explicit xml in query", http.MethodGet, "/fever/?api=xml", "", true},
		{"json value in query", http.MethodGet, "/fever/?api=json", "", false},
		{"empty api value", http.MethodGet, "/fever/?api", "", false},
		{"no api param", http.MethodGet, "/fever/", "", false},
		{"xml in post body", http.MethodPost, "/fever/", "api=xml", true},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			var r *http.Request
			if scenario.body != "" {
				r = httptest.NewRequest(scenario.method, scenario.target, strings.NewReader(scenario.body))
				r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			} else {
				r = httptest.NewRequest(scenario.method, scenario.target, nil)
			}

			if got := wantsXMLResponse(r); got != scenario.expected {
				t.Errorf("wantsXMLResponse(%q) = %t, want %t", scenario.target, got, scenario.expected)
			}
		})
	}
}

func TestSendResponseDefaultsToJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/fever/", nil)
	w := httptest.NewRecorder()

	sendResponse(w, r, sampleGroupsResponse())

	resp := w.Result()
	defer resp.Body.Close()

	if got := resp.Header.Get("Content-Type"); got != "application/json" {
		t.Fatalf("unexpected content type, got %q", got)
	}

	body := w.Body.String()
	if !strings.Contains(body, `"api_version":3`) {
		t.Errorf("JSON body missing api_version: %s", body)
	}
	if !strings.Contains(body, `"groups":[{"id":1,"title":"Tech"}]`) {
		t.Errorf("JSON body missing groups: %s", body)
	}
	// The XMLName field added for XML support must never leak into JSON.
	if strings.Contains(body, "XMLName") {
		t.Errorf("JSON body unexpectedly contains XMLName: %s", body)
	}
}

func TestSendResponseXML(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/fever/?api=xml", nil)
	w := httptest.NewRecorder()

	sendResponse(w, r, sampleGroupsResponse())

	resp := w.Result()
	defer resp.Body.Close()

	if got := resp.Header.Get("Content-Type"); got != "text/xml; charset=utf-8" {
		t.Fatalf("unexpected content type, got %q", got)
	}

	body := w.Body.String()
	if !strings.HasPrefix(body, "<?xml") {
		t.Errorf("XML body missing declaration: %s", body)
	}

	for _, fragment := range []string{
		"<response>",
		"<api_version>3</api_version>",
		"<groups><group><id>1</id><title>Tech</title></group></groups>",
		"<feeds_groups><feeds_group><group_id>1</group_id><feed_ids>10,11</feed_ids></feeds_group></feeds_groups>",
	} {
		if !strings.Contains(body, fragment) {
			t.Errorf("XML body missing %q: %s", fragment, body)
		}
	}
}

func TestAuthFailureResponseFormats(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/fever/", nil)
		w := httptest.NewRecorder()

		sendResponse(w, r, newAuthFailureResponse())

		body := w.Body.String()
		if !strings.Contains(body, `"auth":0`) {
			t.Errorf("JSON auth failure missing auth=0: %s", body)
		}
		if strings.Contains(body, "XMLName") {
			t.Errorf("JSON auth failure contains XMLName: %s", body)
		}
	})

	t.Run("xml", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/fever/?api=xml", nil)
		w := httptest.NewRecorder()

		sendResponse(w, r, newAuthFailureResponse())

		body := w.Body.String()
		if !strings.Contains(body, "<response><api_version>3</api_version><auth>0</auth>") {
			t.Errorf("XML auth failure unexpected body: %s", body)
		}
	})
}

func TestSendServerErrorFormats(t *testing.T) {
	t.Run("json", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/fever/", nil)
		w := httptest.NewRecorder()

		sendServerError(w, r, errors.New("boom"))

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("unexpected status, got %d", resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected content type, got %q", got)
		}
		if body := w.Body.String(); !strings.Contains(body, "boom") {
			t.Errorf("JSON error body missing message: %s", body)
		}
	})

	t.Run("xml", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/fever/?api=xml", nil)
		w := httptest.NewRecorder()

		sendServerError(w, r, errors.New("boom"))

		resp := w.Result()
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("unexpected status, got %d", resp.StatusCode)
		}
		if got := resp.Header.Get("Content-Type"); got != "text/xml; charset=utf-8" {
			t.Fatalf("unexpected content type, got %q", got)
		}
		if body := w.Body.String(); !strings.Contains(body, "<error><error_message>boom</error_message></error>") {
			t.Errorf("XML error body unexpected: %s", body)
		}
	})
}
