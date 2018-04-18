release: clean
	goreleaser --snapshot
.PHONY: clean

clean:
	rm -rf dist
.PHONY: clean