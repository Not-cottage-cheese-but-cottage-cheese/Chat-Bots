package main

import "regexp"

func isAlbumURL(str string) bool {
	re := regexp.MustCompile(`(https?:\/\/)?vk.com\/album(-?\d+)_(\d+)`)
	return re.Match([]byte(str))
}
