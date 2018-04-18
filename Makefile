clean:
	rm -rf dist
.PHONY: clean

release: clean
	goreleaser --snapshot
.PHONY: clean