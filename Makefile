default: static

static:
	docker run --rm -v $(shell pwd):/src centurylink/golang-builder