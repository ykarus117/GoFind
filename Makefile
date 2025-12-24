NAME = GoFind
INSTPATH = /usr/local/bin/
BIN = ./bin
BUILD = go build -tags "sqlite_foreign_keys" -o
WINDOWS = GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc
LINUX = GOOS=linux GOARCH=amd64 CGO_ENABLED=1
ARM64 = GOOS=linux GOARCH=arm64 CGO_ENABLED=1 CC=arm-linux-gnueabihf-gcc CXX=0
USER = gofind_user

.PHONY: prepare uninstall install cleanup cc_build

build:
	CGO_ENABLED=1 $(BUILD) $(BIN)/$(NAME) src/main.go

cc_build:
	$(WINDOWS) $(BUILD) $(BIN)/$(NAME)_windows_x64 src/main.go
	$(LINUX) $(BUILD) $(BIN)/$(NAME)_linux_x64 src/main.go
	$(ARM64) $(BUILD) $(BIN)/$(NAME)_linux_arm64 src/main.go

cleanup:
	rm -rf bin/

uninstall:
	systemctl stop gofind
	systemctl disable gofind
	rm /etc/systemd/system/gofind.service
	rm -rf $(INSTPATH)
	userdel $(USER)
	systemctl daemon-reload

install: build
	@useradd -r -s /bin/false $(USER)
	@cp src/systemd/gofind.service /etc/systemd/system/gofind.service
	@mkdir -p $(INSTPATH)/$(NAME)/front
	@cp $(BIN)/$(NAME) $(INSTPATH)/$(NAME)
	@cp ./front $(INSTPATH)/$(NAME) -r
	@chown -R $(USER):$(USER) $(INSTPATH)/$(NAME)
	@systemctl daemon-reload
	@systemctl enable gofind
	@systemctl start gofind
	@rm -rf $(BIN)
	@echo ================================
	@echo Installed $(NAME) to $(INSTPATH)/$(NAME)
	@echo Run 'systemctl start|stop|restart|status gofind' to manage the service
	@echo Access the web interface at http://localhost:8080