// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"miniflux.app/v2/internal/model"
)

func TestSendEntryLogsLinkwardenCollectionID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	prev := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(prev)

	entry := &model.Entry{ID: 52, URL: "https://example.org/test.html", Title: "Test"}
	coll := int64(12345)
	userIntegrations := &model.Integration{
		UserID:                 1,
		LinkwardenEnabled:      true,
		LinkwardenCollectionID: &coll,
		LinkwardenURL:          "",
		LinkwardenAPIKey:       "",
	}

	SendEntry(entry, userIntegrations)

	out := buf.String()
	if !strings.Contains(out, `"collection_id":12345`) {
		t.Fatalf("expected collection_id in logs; got: %s", out)
	}
}

func TestSendEntryLogsLinkwardenWithoutCollectionID(t *testing.T) {
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(handler)
	prev := slog.Default()
	slog.SetDefault(logger)
	defer slog.SetDefault(prev)

	entry := &model.Entry{ID: 52, URL: "https://example.org/test.html", Title: "Test"}
	userIntegrations := &model.Integration{
		UserID:            1,
		LinkwardenEnabled: true,
		LinkwardenURL:     "",
		LinkwardenAPIKey:  "",
	}

	SendEntry(entry, userIntegrations)

	out := buf.String()
	if strings.Contains(out, "collection_id") {
		t.Fatalf("did not expect collection_id in logs; got: %s", out)
	}
}

// captureLogs redirects the default slog logger to a buffer at debug level for the
// duration of fn and returns everything that was logged. Debug level is required
// because the dispatch skeleton emits its "Sending entry to ..." line at debug.
func captureLogs(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug})
	prev := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(prev)
	fn()
	return buf.String()
}

func TestDispatchSaveEntryDisabledSkips(t *testing.T) {
	entry := &model.Entry{ID: 7, URL: "https://example.org/a", Title: "A"}
	called := false

	out := captureLogs(t, func() {
		dispatchSaveEntry(false, "Example", entry, 1, nil, func() error {
			called = true
			return nil
		})
	})

	if called {
		t.Fatal("send must not be called when the provider is disabled")
	}
	if out != "" {
		t.Fatalf("expected no log output when disabled; got: %s", out)
	}
}

func TestDispatchSaveEntrySuccessLogsDebugOnly(t *testing.T) {
	entry := &model.Entry{ID: 7, URL: "https://example.org/a", Title: "A"}
	calls := 0

	out := captureLogs(t, func() {
		dispatchSaveEntry(true, "Example", entry, 42, nil, func() error {
			calls++
			return nil
		})
	})

	if calls != 1 {
		t.Fatalf("expected send to be called once; got %d", calls)
	}
	if !strings.Contains(out, "Sending entry to Example") {
		t.Fatalf("expected debug log; got: %s", out)
	}
	if strings.Contains(out, "Unable to send entry to Example") {
		t.Fatalf("did not expect an error log on success; got: %s", out)
	}
	for _, want := range []string{`"user_id":42`, `"entry_id":7`, `"entry_url":"https://example.org/a"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %s in debug log; got: %s", want, out)
		}
	}
}

func TestDispatchSaveEntryFailureLogsError(t *testing.T) {
	entry := &model.Entry{ID: 7, URL: "https://example.org/a", Title: "A"}

	out := captureLogs(t, func() {
		dispatchSaveEntry(true, "Example", entry, 42, nil, func() error {
			return errors.New("boom")
		})
	})

	if !strings.Contains(out, "Sending entry to Example") {
		t.Fatalf("expected debug log; got: %s", out)
	}
	if !strings.Contains(out, "Unable to send entry to Example") {
		t.Fatalf("expected error log; got: %s", out)
	}
	if !strings.Contains(out, `"error":"boom"`) {
		t.Fatalf("expected error attribute; got: %s", out)
	}
	for _, want := range []string{`"user_id":42`, `"entry_id":7`, `"entry_url":"https://example.org/a"`} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %s in logs; got: %s", want, out)
		}
	}
}

func TestDispatchSaveEntryIncludesExtraAttrsAfterBase(t *testing.T) {
	entry := &model.Entry{ID: 7, URL: "https://example.org/a", Title: "A"}
	extra := []any{slog.String("user_tags", "news,tech")}

	out := captureLogs(t, func() {
		dispatchSaveEntry(true, "Example", entry, 1, extra, func() error {
			return errors.New("boom")
		})
	})

	if !strings.Contains(out, `"user_tags":"news,tech"`) {
		t.Fatalf("expected extra attribute in logs; got: %s", out)
	}
	// Extra attributes must be emitted after the base triple to stay consistent.
	if i, j := strings.Index(out, `"entry_url"`), strings.Index(out, `"user_tags"`); i == -1 || j == -1 || i > j {
		t.Fatalf("expected user_tags to appear after entry_url; got: %s", out)
	}
}

func TestSendEntryAllDisabledDispatchesNothing(t *testing.T) {
	entry := &model.Entry{ID: 7, URL: "https://example.org/a", Title: "A"}

	out := captureLogs(t, func() {
		SendEntry(entry, &model.Integration{UserID: 1})
	})

	if strings.Contains(out, "Sending entry to") {
		t.Fatalf("expected no dispatch when every provider is disabled; got: %s", out)
	}
}

func TestSendEntryNormalizesLinkwardenDebugName(t *testing.T) {
	entry := &model.Entry{ID: 7, URL: "https://example.org/a", Title: "A"}

	out := captureLogs(t, func() {
		SendEntry(entry, &model.Integration{
			UserID:            1,
			LinkwardenEnabled: true,
			LinkwardenURL:     "",
			LinkwardenAPIKey:  "",
		})
	})

	if !strings.Contains(out, "Sending entry to Linkwarden") {
		t.Fatalf("expected normalized debug name 'Linkwarden'; got: %s", out)
	}
	if strings.Contains(out, "Sending entry to linkwarden") {
		t.Fatalf("debug name should be normalized to 'Linkwarden', not lowercase; got: %s", out)
	}
}
