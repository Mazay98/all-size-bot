# suppress output, run `make XXX V=` to be verbose
V := @
NAME = go.sizebot
OUT_DIR = ./bin
MAIN_PKG = ./cmd/${NAME}
LD_FLAGS = -ldflags "-s -v -w -X 'main.version=${VERSION}' -X 'main.buildTime=${CURRENT_TIME}'"
BUILD_CMD = CGO_ENABLED=1 go build -o ${OUT_DIR}/${NAME} ${LD_FLAGS} ${MAIN_PKG}

.PHONY: vendor
vendor:
	$(V)GOPRIVATE=${VCS}/* go mod tidy
	$(V)GOPRIVATE=${VCS}/* go mod vendor
	$(V)git add vendor go.mod go.sum buf.lock

.PHONY: build
prod:
	@echo BUILDING PRODUCTION $(NAME)
	$(V)${BUILD_CMD}
	@echo DONE