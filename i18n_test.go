package main

import (
	"encoding/json"
	"regexp"
	"testing"
)

func TestNormalizeLang(t *testing.T) {
	for in, want := range map[string]string{"ja": "ja", "ja-JP": "ja", "ja_JP.UTF-8": "ja", "JA": "ja", "en": "en", "en_US.UTF-8": "en", "de": "en", "": "en"} {
		if got := normalizeLang(in); got != want {
			t.Errorf("normalizeLang(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
	if s := loadSettings(); s.Lang != "" {
		t.Errorf("settings before save = %+v", s)
	}
	if err := saveSettings(Settings{Lang: "ja"}); err != nil {
		t.Fatal(err)
	}
	if s := loadSettings(); s.Lang != "ja" {
		t.Errorf("settings after save = %+v", s)
	}
}

func TestMessagesHaveSameKeysInEveryLanguage(t *testing.T) {
	for id := range messages["en"] {
		if messages["ja"][id] == "" {
			t.Errorf("message %q has no Japanese text", id)
		}
	}
	for id := range messages["ja"] {
		if messages["en"][id] == "" {
			t.Errorf("message %q has no English text", id)
		}
	}
}

func TestApplyDocTranslationsUsesOnlyMatchingSource(t *testing.T) {
	orig := docsJaJSON
	t.Cleanup(func() { docsJaJSON = orig })
	b, _ := json.Marshal(map[string]docTranslation{
		"same":    {Src: docHash("Original text.\n\nMore."), Doc: "元の文です。\n\n続きです。"},
		"changed": {Src: docHash("Old text."), Doc: "古い訳です。"},
	})
	docsJaJSON = b

	opts := []Option{
		{Key: "same", Doc: "Original text.\n\nMore."},
		{Key: "changed", Doc: "New text."},
		{Key: "missing", Doc: "No translation."},
	}
	applyDocTranslations(opts)
	if opts[0].DocJa != "元の文です。\n\n続きです。" || opts[0].SummaryJa != "元の文です。" {
		t.Errorf("matching entry = %+v", opts[0])
	}
	if opts[1].DocJa != "" || opts[2].DocJa != "" {
		t.Errorf("changed or missing docs must fall back to English: %+v %+v", opts[1], opts[2])
	}
}

// The bundled translation file must stay well-formed.
func TestBundledTranslations(t *testing.T) {
	var ja map[string]docTranslation
	if err := json.Unmarshal(docsJaJSON, &ja); err != nil {
		t.Fatalf("i18n/docs.ja.json is not valid JSON: %v", err)
	}
	if len(ja) == 0 {
		t.Fatal("i18n/docs.ja.json is empty")
	}
	hash := regexp.MustCompile(`^[0-9a-f]{16}$`)
	for key, tr := range ja {
		if !hash.MatchString(tr.Src) {
			t.Errorf("%s: src %q is not a 16-digit hash", key, tr.Src)
		}
		if tr.Doc == "" {
			t.Errorf("%s: empty translation", key)
		}
	}
}
