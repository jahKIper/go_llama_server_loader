package webui

import "embed"

//go:embed static/index.html static/style.css static/app.js
var staticFS embed.FS

// indexHTML contains the embedded index.html content.
var indexHTML []byte

func init() {
	var err error
	indexHTML, err = staticFS.ReadFile("static/index.html")
	if err != nil {
		panic("failed to read index.html: " + err.Error())
	}
}
