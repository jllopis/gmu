.PHONY: help
.DEFAULT_GOAL := help

include config.mk

export PATH := $(TOOLSDIR):$(PATH)
export GO111MODULE := auto

$(BLDDIR):
	mkdir ${BLDDIR} || true

bin: $(BLDDIR) ## build amd64 binary. OS defaults to host OS. This can be overriden setting 'OS' env var (make bin OS=linux)
	$(shell CGO_ENABLED=0 GOOS=${OS} GOARCH=amd64 \
        go build -ldflags ${LDFLAGS} -a -installsuffix cgo \
        -o ${BLDDIR}/${BINNAME}_${VERSION}_${OS}_amd64.bin cmd/gmucli/*.go \
        && chmod +x ${BLDDIR}/${BINNAME}_${VERSION}_${OS}_amd64.bin \
        )

run: bin ## force rebuild the docker image (even if they haven't changed) and run using mydc.yml
	${BLDDIR}/${BINNAME}_${VERSION}_${OS}_amd64.bin --path . \
	--project-url http://fake.com \
	--project-email usr@fake.com \
	--service-name testme \
	--api-version v1 \
	--protoc-version 3.7.0 \
	new github.com/gmu/testProjectFromMake

clean: ## remove the generated files to start clean but keep the images
	rm -rf $(BLDDIR) | true
	rm -rf testprojectfrommake | true

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(firstword $(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'