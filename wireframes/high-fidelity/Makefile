# Filename: Makefile

00-main-menu.png:

%.png: %.svg
	inkscape --export-type=png --export-filename=$@ $<

.PHONY: clean
clean:
	rm -f *.png
