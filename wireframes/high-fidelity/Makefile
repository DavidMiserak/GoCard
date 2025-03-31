# Filename: Makefile

.PHONY: all
all: 00-main-menu.png 01-deck-selection.png 02-01-study-session-answer-hidden.png 02-01-study-session-answer-revealed.png 03-01-stats-summary.png 03-02-stats-deck.png 03-03-stats-forcast.png

00-main-menu.png:
01-deck-selection.png:
02-01-study-session-answer-hidden.png:
02-01-study-session-answer-revealed.png:
03-01-stats-summary.png:
03-02-stats-deck.png:
03-03-stats-forcast.png:

%.png: %.svg
	inkscape --export-type=png --export-filename=$@ $<

.PHONY: clean
clean:
	rm -f *.png
