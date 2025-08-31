#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2025 Markus Stenberg
#
# Created:       Tue Aug 19 15:53:13 2025 mstenber
# Last modified: Sun Aug 31 09:20:08 2025 mstenber
# Edit time:     5 min
#
#

run: build
	go run . data*

build:
	make -C templates

files-for-ai:
	@echo README.md $(wildcard *.go db/*.go templates/*.templ static/*.css)

upgrade:
	go get -u ./...
	go mod tidy
