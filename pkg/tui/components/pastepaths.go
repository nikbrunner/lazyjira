package components

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var imageExtensions = map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true}

// pastedImages returns the images named by pasted text, as terminals paste
// dropped files: shell-escaped, quoted, or file:// paths separated by
// whitespace. It returns nil unless every token is an existing image file.
func pastedImages(text string) []CreateAttachment {
	tokens := splitPastedPaths(text)
	if len(tokens) == 0 {
		return nil
	}
	images := make([]CreateAttachment, 0, len(tokens))
	for _, tok := range tokens {
		if strings.HasPrefix(tok, "file://") {
			u, err := url.Parse(tok)
			if err != nil {
				return nil
			}
			tok = u.Path
		}
		if !imageExtensions[strings.ToLower(filepath.Ext(tok))] {
			return nil
		}
		info, err := os.Stat(tok)
		if err != nil || !info.Mode().IsRegular() {
			return nil
		}
		images = append(images, CreateAttachment{Path: tok, Name: filepath.Base(tok), Size: info.Size()})
	}
	return images
}

// splitPastedPaths splits text at unescaped, unquoted whitespace, removing
// quotes and backslash escapes.
func splitPastedPaths(text string) []string {
	var tokens []string
	var cur strings.Builder
	inToken := false
	var quote rune
	escaped := false
	for _, r := range text {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\\':
			escaped = true
			inToken = true
		case r == '\'' || r == '"':
			quote = r
			inToken = true
		case unicode.IsSpace(r):
			if inToken {
				tokens = append(tokens, cur.String())
				cur.Reset()
				inToken = false
			}
		default:
			cur.WriteRune(r)
			inToken = true
		}
	}
	if inToken {
		tokens = append(tokens, cur.String())
	}
	return tokens
}
