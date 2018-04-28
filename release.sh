#!/bin/bash

if [ ! -n "$1" ]; then
    echo "Error:release version is blank!"
	exit 1
fi

gox -osarch="darwin/amd64 linux/386 linux/amd64" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}"
ghr -u mritd -t $GITHUB_RELEASE_TOKEN -replace -recreate --debug $1 dist/
