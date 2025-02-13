// Copyright 2020, Shulhan <ms@kilabit.info>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bot

// List of message entity types.
const (
	EntityTypeMention       = "mention"       // @username
	EntityTypeHashtag       = "hashtag"       // #hashtag
	EntityTypeBotCommand    = "bot_command"   // /start@jobs_bot
	EntityTypeURL           = "url"           // https://x.y
	EntityTypeEmail         = "email"         // a@b.c
	EntityTypePhoneNumber   = "phone_number"  //+1-234
	EntityTypeBold          = "bold"          // bold text
	EntityTypeItalic        = "italic"        // italic text
	EntityTypeUnderline     = "underline"     // underlined text
	EntityTypeStrikethrough = "strikethrough" // strikethrough text
	EntityTypeCode          = "code"          // monowidth string
	EntityTypePre           = "pre"           // monowidth block
	EntityTypeTextLink      = "text_link"     // for clickable text URLs
	EntityTypeTextMention   = "text_mention"  // for users without usernames.
)

// MessageEntity represents one special entity in a text message. For example,
// hashtags, usernames, URLs, etc.
type MessageEntity struct {
	// Type of the entity.
	Type string `json:"type"`

	// Offset in UTF-16 code units to the start of the entity.
	Offset int `json:"offset"`

	// Length of the entity in UTF-16 code units.
	Length int `json:"length"`

	// Optional. For “text_link” only, url that will be opened after user
	// taps on the text.
	URL string `json:"url"`

	// Optional. For “text_mention” only, the mentioned user.
	User *User `json:"user"`

	// Optional. For “pre” only, the programming language of the entity
	// text.
	Language string `json:"language"`
}
