package utils

import (
	"slices"
	"strings"
)

func IsValidSpotifyURL(url string) bool {
	// Check if the URL starts with "https://open.spotify.com/album/"
	return strings.HasPrefix(url, "https://open.spotify.com/")
}

func InWhiteList(url int64, whitelist []int64) bool {
	return slices.Contains(whitelist, url)
}
