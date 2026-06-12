// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package form // import "miniflux.app/v2/internal/ui/form"

import (
	"reflect"
	"testing"
)

// newFullySetSubscriptionForm returns a SubscriptionForm with every field set
// to a distinct, non-zero value so the mapping to a FeedCreationRequest can be
// verified field by field.
func newFullySetSubscriptionForm() *SubscriptionForm {
	return &SubscriptionForm{
		URL:                         "https://example.org/site",
		CategoryID:                  42,
		Crawler:                     true,
		IgnoreEntryUpdates:          true,
		FetchViaProxy:               true,
		AllowSelfSignedCertificates: true,
		UserAgent:                   "Custom UA",
		Cookie:                      "session=abc",
		Username:                    "feed-user",
		Password:                    "feed-secret",
		ScraperRules:                "//article",
		RewriteRules:                "add_dynamic_image",
		UrlRewriteRules:             `rewrite("title")`,
		BlocklistRules:              "spam",
		KeeplistRules:               "news",
		BlockFilterEntryRules:       "EntryTitle=spam",
		KeepFilterEntryRules:        "EntryTitle=news",
		DisableHTTP2:                true,
		ProxyURL:                    "http://proxy.example.org",
	}
}

func TestSubscriptionFormNewFeedCreationRequest(t *testing.T) {
	subscriptionForm := newFullySetSubscriptionForm()

	const feedURL = "https://example.org/feed.xml"
	request := subscriptionForm.NewFeedCreationRequest(feedURL)

	// The feed URL must come from the argument (the discovered or chosen feed
	// URL), not from the URL the user submitted in the form.
	if request.FeedURL != feedURL {
		t.Errorf("FeedURL = %q, want %q", request.FeedURL, feedURL)
	}
	if request.FeedURL == subscriptionForm.URL {
		t.Errorf("FeedURL must not fall back to the form URL %q", subscriptionForm.URL)
	}

	assertEqual := func(name string, got, want any) {
		t.Helper()
		if got != want {
			t.Errorf("%s = %v, want %v", name, got, want)
		}
	}

	assertEqual("CategoryID", request.CategoryID, subscriptionForm.CategoryID)
	assertEqual("Crawler", request.Crawler, subscriptionForm.Crawler)
	assertEqual("IgnoreEntryUpdates", request.IgnoreEntryUpdates, subscriptionForm.IgnoreEntryUpdates)
	assertEqual("FetchViaProxy", request.FetchViaProxy, subscriptionForm.FetchViaProxy)
	assertEqual("AllowSelfSignedCertificates", request.AllowSelfSignedCertificates, subscriptionForm.AllowSelfSignedCertificates)
	assertEqual("UserAgent", request.UserAgent, subscriptionForm.UserAgent)
	assertEqual("Cookie", request.Cookie, subscriptionForm.Cookie)
	assertEqual("Username", request.Username, subscriptionForm.Username)
	assertEqual("Password", request.Password, subscriptionForm.Password)
	assertEqual("ScraperRules", request.ScraperRules, subscriptionForm.ScraperRules)
	assertEqual("RewriteRules", request.RewriteRules, subscriptionForm.RewriteRules)
	assertEqual("UrlRewriteRules", request.UrlRewriteRules, subscriptionForm.UrlRewriteRules)
	assertEqual("BlocklistRules", request.BlocklistRules, subscriptionForm.BlocklistRules)
	assertEqual("KeeplistRules", request.KeeplistRules, subscriptionForm.KeeplistRules)
	assertEqual("BlockFilterEntryRules", request.BlockFilterEntryRules, subscriptionForm.BlockFilterEntryRules)
	assertEqual("KeepFilterEntryRules", request.KeepFilterEntryRules, subscriptionForm.KeepFilterEntryRules)
	assertEqual("DisableHTTP2", request.DisableHTTP2, subscriptionForm.DisableHTTP2)
	assertEqual("ProxyURL", request.ProxyURL, subscriptionForm.ProxyURL)

	// Fields that the subscription form does not control must stay at their
	// zero value rather than being silently populated.
	if request.Disabled || request.NoMediaPlayer || request.IgnoreHTTPCache || request.HideGlobally {
		t.Errorf("a field not owned by the subscription form was set: %+v", request)
	}
}

// TestSubscriptionFormNewFeedCreationRequestPropagatesEveryField guards against
// drift: if a field is added to SubscriptionForm but not propagated by
// NewFeedCreationRequest, this test fails. That keeps proxy, authentication and
// custom-rule settings consistent across every "add subscription" entry point.
func TestSubscriptionFormNewFeedCreationRequestPropagatesEveryField(t *testing.T) {
	subscriptionForm := newFullySetSubscriptionForm()

	// Align the feed URL with the form URL so the generic comparison below can
	// treat URL/FeedURL like any other field.
	request := subscriptionForm.NewFeedCreationRequest(subscriptionForm.URL)

	// URL is the only field renamed on the request side; every other field must
	// share the same name on both structs.
	requestFieldName := func(formFieldName string) string {
		if formFieldName == "URL" {
			return "FeedURL"
		}
		return formFieldName
	}

	formValue := reflect.ValueOf(*subscriptionForm)
	formType := formValue.Type()
	requestValue := reflect.ValueOf(*request)

	for i := 0; i < formType.NumField(); i++ {
		formFieldName := formType.Field(i).Name
		want := formValue.Field(i)

		// A zero value in the fixture would make this guard meaningless, so
		// require newFullySetSubscriptionForm to populate every field.
		if want.IsZero() {
			t.Fatalf("newFullySetSubscriptionForm leaves SubscriptionForm.%s at its zero value; set it so the mapping can be verified", formFieldName)
		}

		requestField := requestValue.FieldByName(requestFieldName(formFieldName))
		if !requestField.IsValid() {
			t.Errorf("SubscriptionForm.%s has no counterpart on FeedCreationRequest", formFieldName)
			continue
		}

		if !reflect.DeepEqual(requestField.Interface(), want.Interface()) {
			t.Errorf("SubscriptionForm.%s is not propagated to FeedCreationRequest: got %v, want %v", formFieldName, requestField.Interface(), want.Interface())
		}
	}
}
