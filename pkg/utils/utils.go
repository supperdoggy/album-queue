package utils

import "strings"

func IsValidSpotifyURL(url string) bool {
	// Check if the URL starts with "https://open.spotify.com/album/"
	return strings.HasPrefix(url, "https://open.spotify.com/")
}
