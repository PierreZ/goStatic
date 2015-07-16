default: static

static:
	docker run --rm -v $(pwd):/src centurylink/golang-builder