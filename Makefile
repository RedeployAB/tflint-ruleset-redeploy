default: build

test:
	go test --count=1 $$(go list ./... | grep -v integration)

build:
ifeq ($(OS),Windows_NT)
	go build -o tflint-ruleset-redeploy.exe
else
	go build -o tflint-ruleset-redeploy
endif

install: build
ifeq ($(OS),Windows_NT)
	mkdir -p $(USERPROFILE)/.tflint.d/plugins
	mv tflint-ruleset-redeploy.exe $(USERPROFILE)/.tflint.d/plugins
else
	mkdir -p ~/.tflint.d/plugins
	mv ./tflint-ruleset-redeploy ~/.tflint.d/plugins
endif

e2e: install
	cd integration && go test -v && cd ../

.PHONY: test build install e2e
