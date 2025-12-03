ARCH?=amd64
REPO?=#your repository here
VERSION?=0.1

build:
	CGO_ENABLED=0 GOOS=linux GOARCH=$(ARCH) go build -o ./bin/prometheus-scraper main.go

container:
	docker build -t $(REPO)finops-prometheus-scraper:$(VERSION) .
	docker push $(REPO)finops-prometheus-scraper:$(VERSION)

container-multi:
	docker buildx build --tag $(REPO)finops-prometheus-scraper:$(VERSION) --push --platform linux/amd64,linux/arm64 .