
# Simple go build
postgres_exporter: postgres_exporter.go
	go build -o postgres_exporter .

# Do a self-contained docker build - we pull the official upstream container,
# then template out a dockerfile which builds the real image.
docker:
	docker run -v $(shell pwd):/go/src/github.com/wrouesnel/postgres_exporter \
		golang:1.6-wheezy \
		/go/src/github.com/wrouesnel/postgres_exporter/docker-build.bsh /postgres_exporter /go/src/github.com/wrouesnel/postgres_exporter | \
		docker import --change "EXPOSE 9113" \
			--change 'ENTRYPOINT [ "/postgres_exporter" ]' \
			- wrouesnel/postgres_exporter

.PHONY: docker
