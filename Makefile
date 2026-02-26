#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2025 Markus Stenberg
#
# Created:       Tue Aug 19 15:53:13 2025 mstenber
# Last modified: Sat Feb 21 18:30:13 2026 mstenber
# Edit time:     25 min
#
#

run: build
	go run . data*

build: lint

lint: templates
	go tool golangci-lint run

test:
	go test ./...

templates:
	make -C templates

upgrade:
	go get -u ./...
	go mod tidy

# Semi-convenience - can be used to bump tools
install-tools:
	go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go get -tool github.com/a-h/templ/cmd/templ@latest

# This should be run as root to get container working
dep-debian-13:
	# prek needs git
	apt-get install -y git golang pipx
	pipx install uv
	pipx ensurepath
	bash -c 'source ~/.bashrc && uv tool install prek'

# This is cecli/aider crutch
files-for-ai:
	@echo README.md $(wildcard *.go db/*.go templates/*.templ static/*.css)
