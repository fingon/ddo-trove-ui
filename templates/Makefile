#
# Author: Markus Stenberg <fingon@iki.fi>
#
# Copyright (c) 2025 Markus Stenberg
#
# Created:       Tue Aug 19 15:56:42 2025 mstenber
# Last modified: Tue Aug 19 16:05:07 2025 mstenber
# Edit time:     2 min
#
#

TEMPLATES = $(wildcard *.templ)
GENERATED = $(patsubst %.templ,%_templ.go,$(TEMPLATES))

all: $(GENERATED)

%_templ.go: %.templ Makefile
	go tool templ generate -f $<
