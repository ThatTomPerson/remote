release: clean
	goreleaser
	snapcraft push dist/remote_*_linux_amd64.snap
.PHONY: clean

clean:
	rm -rf dist
.PHONY: clean