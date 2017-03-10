
lint:
	gometalinter \
		--vendor \
		--deadline=60s \
		--disable-all \
		--enable=vet \
		--enable=golint \
		--enable=gosimple \
		--enable=gofmt \
		--enable=misspell \
		./...
