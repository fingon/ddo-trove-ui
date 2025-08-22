#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2025 Markus Stenberg
#
# Created:       Tue Aug 19 15:53:13 2025 mstenber
# Last modified: Thu Aug 21 14:35:24 2025 mstenber
# Edit time:     4 min
#
#

run: build
	go run . --input data

build:
	make -C templates

files-for-ai:
	@echo README.md $(wildcard *.go db/*.go templates/*.templ static/*.css)

upgrade:
	go get -u ./...
	go mod tidy
