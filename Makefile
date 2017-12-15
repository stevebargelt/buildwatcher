SOURCEDIR=.
SOURCES = $(shell find $(SOURCEDIR) -name '*.go')
VERSION=$(shell git describe --always --tags)
BINARY=bin/buildwatcher

.PHONY:pi
pi:
	env GOOS=linux GOARCH=arm go build -o $(BINARY) -ldflags "-s -w -X main.Version=$(VERSION)" main.go
	cp buildwatcher.yml bin/config.yml	

.PHONY:pi-zero
pi-zero:
	env GOARM=6 GOOS=linux GOARCH=arm go build -o $(BINARY) -ldflags "-s -w -X main.Version=$(VERSION)" commands/*.go

.PHONY:deploy
deploy:
	scp bin/buildwatcher pi@192.168.1.42:buildwatcher
	#scp bin/config.yml pi@192.168.1.42:config.yml
test:
	go test -cover -v -race ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: docker
docker:
	docker build -t buildwatcher -f Dockerfile .

# docker run --name buildwatcher -it example-scratch


# .PHONY: build
# build: clean go-get test bin

# .PHONY: deb
# deb:
# 	mkdir -p dist/var/lib/buildwatcher/assets dist/usr/bin dist/etc/buildwatcher
# 	cp bin/buildwatcher dist/usr/bin/buildwatcher
# 	cp build/buildwatcher.yml dist/etc/buildwatcher/config.yml
# 	bundle exec fpm -t deb -s dir -a armhf -n buildwatcher -v $(VERSION) -m steve@bargelt.com --deb-systemd build/buildwatcher.service -C dist  -p buildwatcher-$(VERSION).deb .

 .PHONY: clean
 clean:
	-rm -rf *.deb
	-rm -rf dist
	-rm *.db
