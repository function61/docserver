#!/bin/bash -eu

run() {
	fn="$1"

	echo "# $fn"

	"$fn"
}

downloadDependencies() {
	dep ensure
}

checkFormatting() {
	# unfortunately we need to list formattable directories because "." would include vendor/
	local offenders=$(ls *.go | xargs gofmt -l)

	if [ ! -z "$offenders" ]; then
		>&2 echo "formatting errors: $offenders"
		exit 1
	fi
}

unitTests() {
	go test ./...
}

staticAnalysis() {
	go vet ./...
}

buildLinuxAmd64() {
	# compile statically so this works on Alpine that doesn't have glibc
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$FRIENDLY_REV_ID -extldflags \"-static\"" -o docserver
}

rm -rf rel
mkdir rel

run downloadDependencies

run checkFormatting

run staticAnalysis

run unitTests

run buildLinuxAmd64
