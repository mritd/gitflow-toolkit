BUILD_VERSION   	:= $(shell cat version)
BUILD_DATE      	:= $(shell date "+%F %T")
COMMIT_SHA1     	:= $(shell git rev-parse HEAD)

all: clean
	bash .cross_compile.sh

release: all
	ghr -u mritd -t ${GITHUB_TOKEN} -replace -recreate -name "Bump ${BUILD_VERSION}" --debug ${BUILD_VERSION} dist

pre-release: all
	ghr -u mritd -t ${GITHUB_TOKEN} -replace -recreate -prerelease -name "Bump ${BUILD_VERSION}" --debug ${BUILD_VERSION} dist

install:
	go install -trimpath -ldflags	"-X 'main.version=${BUILD_VERSION}' \
               						-X 'main.buildDate=${BUILD_DATE}' \
               						-X 'main.commitID=${COMMIT_SHA1}'"
debug:
	go install  -trimpath -gcflags "all=-N -l"

clean:
	rm -rf dist

.PHONY: all release pre-release clean install debug
