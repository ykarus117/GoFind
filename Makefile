NAME = GoFind
PATH = /usr/local/bin/$(NAME)
BIN = bin/
BUILD = go build -tags "sqlite_foreign_keys" -o
.PHONY: prepare uninstall

cc_build:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 $(BUILD) $(BIN)/$(NAME)_windows_x64 src/main.go
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=1 $(BUILD) $(BIN)/$(NAME)_linux_x64 src/main.go
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=1  $(BUILD) $(BIN)/$(NAME)_linux_arm64 src/main.go
	GOOS=windows GOARCH=arm64 CGO_ENABLED=1  $(BUILD) $(BIN)/$(NAME)_windows_arm64 src/main.go

build:
	CGO_ENABLED=1 $(BUILD) /usr/local/bin/$(NAME) src/main.go

prepare:
	mkdir $(PATH)
	mkdir $(PATH)/front
	cp src/front/* $(PATH)/front -r

uninstall:
	service gofind stop
	service gofind disable
	rm /etc/systemd/system/gofind.service
	rm -rf $(PATH)

install: build | prepare
	cp src/systemd/gofind.service /etc/systemd/system/gofind.service
	systemctl daemon-reload
	service gofind enable
	service gofind start