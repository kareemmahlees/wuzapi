package main

import (
	"encoding/base64"

	"github.com/skip2/go-qrcode"
)

func ToBase64Image(input string) string {
	image, _ := qrcode.Encode(input, qrcode.Medium, 256)
	base64qrcode := "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)
	return base64qrcode
}
