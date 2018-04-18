release: clean
	goreleaser
	snapcraft push dist/dps-remote_*.snap
.PHONY: clean

clean:
	rm -rf dist
.PHONY: clean