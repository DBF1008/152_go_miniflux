// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package integration // import "miniflux.app/v2/internal/integration"

import (
	"log/slog"

	"miniflux.app/v2/internal/integration/apprise"
	"miniflux.app/v2/internal/integration/archiveorg"
	"miniflux.app/v2/internal/integration/betula"
	"miniflux.app/v2/internal/integration/cubox"
	"miniflux.app/v2/internal/integration/discord"
	"miniflux.app/v2/internal/integration/espial"
	"miniflux.app/v2/internal/integration/instapaper"
	"miniflux.app/v2/internal/integration/karakeep"
	"miniflux.app/v2/internal/integration/linkace"
	"miniflux.app/v2/internal/integration/linkding"
	"miniflux.app/v2/internal/integration/linktaco"
	"miniflux.app/v2/internal/integration/linkwarden"
	"miniflux.app/v2/internal/integration/matrixbot"
	"miniflux.app/v2/internal/integration/notion"
	"miniflux.app/v2/internal/integration/ntfy"
	"miniflux.app/v2/internal/integration/nunuxkeeper"
	"miniflux.app/v2/internal/integration/omnivore"
	"miniflux.app/v2/internal/integration/pinboard"
	"miniflux.app/v2/internal/integration/pushover"
	"miniflux.app/v2/internal/integration/raindrop"
	"miniflux.app/v2/internal/integration/readeck"
	"miniflux.app/v2/internal/integration/readwise"
	"miniflux.app/v2/internal/integration/shaarli"
	"miniflux.app/v2/internal/integration/shiori"
	"miniflux.app/v2/internal/integration/slack"
	"miniflux.app/v2/internal/integration/telegrambot"
	"miniflux.app/v2/internal/integration/wallabag"
	"miniflux.app/v2/internal/integration/webhook"
	"miniflux.app/v2/internal/model"
)

// dispatchSaveEntry runs a single "save entry" provider in a uniform way: when the
// provider is enabled it logs a debug line, performs the send, and logs an error
// when the send fails. The base log attributes (user_id, entry_id, entry_url) are
// always emitted first, followed by any provider-specific attributes. Keeping this
// skeleton in one place ensures logging and provider invocation stay consistent as
// providers are added.
func dispatchSaveEntry(enabled bool, provider string, entry *model.Entry, userID int64, extraAttrs []any, send func() error) {
	if !enabled {
		return
	}

	attrs := append([]any{
		slog.Int64("user_id", userID),
		slog.Int64("entry_id", entry.ID),
		slog.String("entry_url", entry.URL),
	}, extraAttrs...)

	slog.Debug("Sending entry to "+provider, attrs...)

	if err := send(); err != nil {
		slog.Error("Unable to send entry to "+provider, append(attrs, slog.Any("error", err))...)
	}
}

// SendEntry sends the entry to third-party providers when the user click on "Save".
func SendEntry(entry *model.Entry, userIntegrations *model.Integration) {
	dispatchSaveEntry(userIntegrations.BetulaEnabled, "Betula", entry, userIntegrations.UserID, nil, func() error {
		return betula.NewClient(userIntegrations.BetulaURL, userIntegrations.BetulaToken).CreateBookmark(
			entry.URL,
			entry.Title,
			entry.Tags,
		)
	})

	dispatchSaveEntry(userIntegrations.PinboardEnabled, "Pinboard", entry, userIntegrations.UserID, nil, func() error {
		return pinboard.NewClient(userIntegrations.PinboardToken).CreateBookmark(
			entry.URL,
			entry.Title,
			userIntegrations.PinboardTags,
			userIntegrations.PinboardMarkAsUnread,
		)
	})

	dispatchSaveEntry(userIntegrations.InstapaperEnabled, "Instapaper", entry, userIntegrations.UserID, nil, func() error {
		return instapaper.NewClient(userIntegrations.InstapaperUsername, userIntegrations.InstapaperPassword).AddURL(entry.URL, entry.Title)
	})

	dispatchSaveEntry(userIntegrations.WallabagEnabled, "Wallabag", entry, userIntegrations.UserID,
		[]any{slog.String("user_tags", userIntegrations.WallabagTags)},
		func() error {
			return wallabag.NewClient(
				userIntegrations.WallabagURL,
				userIntegrations.WallabagClientID,
				userIntegrations.WallabagClientSecret,
				userIntegrations.WallabagUsername,
				userIntegrations.WallabagPassword,
				userIntegrations.WallabagTags,
				userIntegrations.WallabagOnlyURL,
			).CreateEntry(entry.URL, entry.Title, entry.Content)
		})

	dispatchSaveEntry(userIntegrations.NotionEnabled, "Notion", entry, userIntegrations.UserID, nil, func() error {
		return notion.NewClient(
			userIntegrations.NotionToken,
			userIntegrations.NotionPageID,
		).UpdateDocument(entry.URL, entry.Title)
	})

	dispatchSaveEntry(userIntegrations.NunuxKeeperEnabled, "NunuxKeeper", entry, userIntegrations.UserID, nil, func() error {
		return nunuxkeeper.NewClient(
			userIntegrations.NunuxKeeperURL,
			userIntegrations.NunuxKeeperAPIKey,
		).AddEntry(entry.URL, entry.Title, entry.Content)
	})

	dispatchSaveEntry(userIntegrations.EspialEnabled, "Espial", entry, userIntegrations.UserID, nil, func() error {
		return espial.NewClient(
			userIntegrations.EspialURL,
			userIntegrations.EspialAPIKey,
		).CreateLink(entry.URL, entry.Title, userIntegrations.EspialTags)
	})

	dispatchSaveEntry(userIntegrations.LinkAceEnabled, "LinkAce", entry, userIntegrations.UserID, nil, func() error {
		return linkace.NewClient(
			userIntegrations.LinkAceURL,
			userIntegrations.LinkAceAPIKey,
			userIntegrations.LinkAceTags,
			userIntegrations.LinkAcePrivate,
			userIntegrations.LinkAceCheckDisabled,
		).AddURL(entry.URL, entry.Title)
	})

	dispatchSaveEntry(userIntegrations.LinkdingEnabled, "Linkding", entry, userIntegrations.UserID, nil, func() error {
		return linkding.NewClient(
			userIntegrations.LinkdingURL,
			userIntegrations.LinkdingAPIKey,
			userIntegrations.LinkdingTags,
			userIntegrations.LinkdingMarkAsUnread,
		).CreateBookmark(entry.URL, entry.Title)
	})

	dispatchSaveEntry(userIntegrations.LinktacoEnabled, "LinkTaco", entry, userIntegrations.UserID, nil, func() error {
		return linktaco.NewClient(
			userIntegrations.LinktacoAPIToken,
			userIntegrations.LinktacoOrgSlug,
			userIntegrations.LinktacoTags,
			userIntegrations.LinktacoVisibility,
		).CreateBookmark(entry.URL, entry.Title, entry.Content)
	})

	var linkwardenExtra []any
	if userIntegrations.LinkwardenCollectionID != nil {
		linkwardenExtra = []any{slog.Int64("collection_id", *userIntegrations.LinkwardenCollectionID)}
	}
	dispatchSaveEntry(userIntegrations.LinkwardenEnabled, "Linkwarden", entry, userIntegrations.UserID, linkwardenExtra, func() error {
		return linkwarden.NewClient(
			userIntegrations.LinkwardenURL,
			userIntegrations.LinkwardenAPIKey,
			userIntegrations.LinkwardenCollectionID,
		).CreateBookmark(entry.URL, entry.Title)
	})

	dispatchSaveEntry(userIntegrations.ReadeckEnabled, "Readeck", entry, userIntegrations.UserID, nil, func() error {
		return readeck.NewClient(
			userIntegrations.ReadeckURL,
			userIntegrations.ReadeckAPIKey,
			userIntegrations.ReadeckLabels,
			userIntegrations.ReadeckOnlyURL,
		).CreateBookmark(entry.URL, entry.Title, entry.Content)
	})

	dispatchSaveEntry(userIntegrations.ReadwiseEnabled, "Readwise", entry, userIntegrations.UserID, nil, func() error {
		return readwise.NewClient(
			userIntegrations.ReadwiseAPIKey,
		).CreateDocument(entry.URL)
	})

	dispatchSaveEntry(userIntegrations.CuboxEnabled, "Cubox", entry, userIntegrations.UserID, nil, func() error {
		return cubox.NewClient(userIntegrations.CuboxAPILink).SaveLink(entry.URL)
	})

	dispatchSaveEntry(userIntegrations.ShioriEnabled, "Shiori", entry, userIntegrations.UserID, nil, func() error {
		return shiori.NewClient(
			userIntegrations.ShioriURL,
			userIntegrations.ShioriUsername,
			userIntegrations.ShioriPassword,
		).CreateBookmark(entry.URL, entry.Title)
	})

	dispatchSaveEntry(userIntegrations.ShaarliEnabled, "Shaarli", entry, userIntegrations.UserID, nil, func() error {
		return shaarli.NewClient(
			userIntegrations.ShaarliURL,
			userIntegrations.ShaarliAPISecret,
		).CreateLink(entry.URL, entry.Title)
	})

	dispatchSaveEntry(userIntegrations.ArchiveorgEnabled, "Archive.org", entry, userIntegrations.UserID, nil, func() error {
		return archiveorg.NewClient().SendURL(entry.URL)
	})

	webhookURL := userIntegrations.WebhookURL
	if entry.Feed != nil && entry.Feed.WebhookURL != "" {
		webhookURL = entry.Feed.WebhookURL
	}
	dispatchSaveEntry(userIntegrations.WebhookEnabled, "Webhook", entry, userIntegrations.UserID,
		[]any{slog.String("webhook_url", webhookURL)},
		func() error {
			return webhook.NewClient(webhookURL, userIntegrations.WebhookSecret).SendSaveEntryWebhookEvent(entry)
		})

	dispatchSaveEntry(userIntegrations.OmnivoreEnabled, "Omnivore", entry, userIntegrations.UserID, nil, func() error {
		return omnivore.NewClient(userIntegrations.OmnivoreAPIKey, userIntegrations.OmnivoreURL).SaveURL(entry.URL)
	})

	dispatchSaveEntry(userIntegrations.KarakeepEnabled, "Karakeep", entry, userIntegrations.UserID,
		[]any{slog.String("user_tags", userIntegrations.KarakeepTags)},
		func() error {
			return karakeep.NewClient(
				userIntegrations.KarakeepAPIKey,
				userIntegrations.KarakeepURL,
				userIntegrations.KarakeepTags,
			).SaveURL(entry.URL)
		})

	dispatchSaveEntry(userIntegrations.RaindropEnabled, "Raindrop", entry, userIntegrations.UserID, nil, func() error {
		return raindrop.NewClient(userIntegrations.RaindropToken, userIntegrations.RaindropCollectionID, userIntegrations.RaindropTags).CreateRaindrop(entry.URL, entry.Title)
	})
}

// PushEntries pushes a list of entries to activated third-party providers during feed refreshes.
func PushEntries(feed *model.Feed, entries model.Entries, userIntegrations *model.Integration) {
	if userIntegrations.MatrixBotEnabled {
		slog.Debug("Sending new entries to Matrix",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		err := matrixbot.PushEntries(
			feed,
			entries,
			userIntegrations.MatrixBotURL,
			userIntegrations.MatrixBotUser,
			userIntegrations.MatrixBotPassword,
			userIntegrations.MatrixBotChatID,
		)
		if err != nil {
			slog.Error("Unable to send new entries to Matrix",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int("nb_entries", len(entries)),
				slog.Int64("feed_id", feed.ID),
				slog.Any("error", err),
			)
		}
	}
	if userIntegrations.WebhookEnabled {
		var webhookURL string
		if feed.WebhookURL != "" {
			webhookURL = feed.WebhookURL
		} else {
			webhookURL = userIntegrations.WebhookURL
		}

		slog.Debug("Sending new entries to Webhook",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
			slog.String("webhook_url", webhookURL),
		)

		webhookClient := webhook.NewClient(webhookURL, userIntegrations.WebhookSecret)
		if err := webhookClient.SendNewEntriesWebhookEvent(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Webhook",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int("nb_entries", len(entries)),
				slog.Int64("feed_id", feed.ID),
				slog.String("webhook_url", webhookURL),
				slog.Any("error", err),
			)
		}
	}

	if userIntegrations.NtfyEnabled && feed.NtfyEnabled {
		ntfyTopic := feed.NtfyTopic
		if ntfyTopic == "" {
			ntfyTopic = userIntegrations.NtfyTopic
		}
		slog.Debug("Sending new entries to Ntfy",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
			slog.String("topic", ntfyTopic),
		)

		client := ntfy.NewClient(
			userIntegrations.NtfyURL,
			ntfyTopic,
			userIntegrations.NtfyAPIToken,
			userIntegrations.NtfyUsername,
			userIntegrations.NtfyPassword,
			userIntegrations.NtfyIconURL,
			userIntegrations.NtfyInternalLinks,
			feed.NtfyPriority,
		)

		if err := client.SendMessages(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Ntfy", slog.Any("error", err))
		}
	}

	if userIntegrations.AppriseEnabled {
		slog.Debug("Sending new entries to Apprise",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		appriseServiceURLs := userIntegrations.AppriseServicesURL
		if feed.AppriseServiceURLs != "" {
			appriseServiceURLs = feed.AppriseServiceURLs
		}

		client := apprise.NewClient(
			appriseServiceURLs,
			userIntegrations.AppriseURL,
		)

		if err := client.SendNotification(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Apprise", slog.Any("error", err))
		}
	}

	if userIntegrations.DiscordEnabled {
		slog.Debug("Sending new entries to Discord",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		client := discord.NewClient(
			userIntegrations.DiscordWebhookLink,
		)

		if err := client.SendDiscordMsg(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Discord", slog.Any("error", err))
		}
	}

	if userIntegrations.SlackEnabled {
		slog.Debug("Sending new entries to Slack",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		client := slack.NewClient(
			userIntegrations.SlackWebhookLink,
		)

		if err := client.SendSlackMsg(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Slack", slog.Any("error", err))
		}
	}

	if userIntegrations.PushoverEnabled && feed.PushoverEnabled {
		slog.Debug("Sending new entries to Pushover",
			slog.Int64("user_id", userIntegrations.UserID),
			slog.Int("nb_entries", len(entries)),
			slog.Int64("feed_id", feed.ID),
		)

		client := pushover.NewClient(
			userIntegrations.PushoverUser,
			userIntegrations.PushoverToken,
			feed.PushoverPriority,
			userIntegrations.PushoverDevice,
			userIntegrations.PushoverPrefix,
		)

		if err := client.SendMessages(feed, entries); err != nil {
			slog.Warn("Unable to send new entries to Pushover", slog.Any("error", err))
		}
	}

	// Integrations that only support sending individual entries
	if userIntegrations.TelegramBotEnabled {
		for _, entry := range entries {
			slog.Debug("Sending a new entry to Telegram",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
			)

			if err := telegrambot.PushEntry(
				feed,
				entry,
				userIntegrations.TelegramBotToken,
				userIntegrations.TelegramBotChatID,
				userIntegrations.TelegramBotTopicID,
				userIntegrations.TelegramBotDisableWebPagePreview,
				userIntegrations.TelegramBotDisableNotification,
				userIntegrations.TelegramBotDisableButtons,
			); err != nil {
				slog.Error("Unable to send entry to Telegram",
					slog.Int64("user_id", userIntegrations.UserID),
					slog.Int64("entry_id", entry.ID),
					slog.String("entry_url", entry.URL),
					slog.Any("error", err),
				)
			}
		}
	}

	// Push each new entry to Readeck when push is enabled
	if userIntegrations.ReadeckPushEnabled {
		client := readeck.NewClient(
			userIntegrations.ReadeckURL,
			userIntegrations.ReadeckAPIKey,
			userIntegrations.ReadeckLabels,
			userIntegrations.ReadeckOnlyURL,
		)
		for _, entry := range entries {
			slog.Debug("Sending a new entry to Readeck",
				slog.Int64("user_id", userIntegrations.UserID),
				slog.Int64("entry_id", entry.ID),
				slog.String("entry_url", entry.URL),
			)

			if err := client.CreateBookmark(entry.URL, entry.Title, entry.Content); err != nil {
				slog.Error("Unable to send entry to Readeck",
					slog.Int64("user_id", userIntegrations.UserID),
					slog.Int64("entry_id", entry.ID),
					slog.String("entry_url", entry.URL),
					slog.Any("error", err),
				)
			}
		}
	}
}
