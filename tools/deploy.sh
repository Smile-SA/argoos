#!/bin/bash

# no tag: stop !
TAG=$2
[ -z "$TAG" ] && error "You must pass tag name"

# Write the token in ~/.config/argoos-github-token to be able to
# - create release
# - upload assets
ARGOOS_GIT_TOKEN=$(cat ~/.config/argoos-github-token)

function error() {
    echo $1
    exit 1
}

# create a release
function createRelease() {
    curl -X POST                                                       \
        -H 'Authorization: token '${ARGOOS_GIT_TOKEN}                  \
        -H 'Accept: application/vnd.github.v3+json'                    \
        -d '{"tag_name": "v1-beta3", "name": "'$TAG'", "draft": true}' \
        https://api.github.com/repos/Smile-SA/argoos/releases
}

# get url to upload binary as asset for that release
function getUploadURL() {
    echo "Get upload..." >&2
    curl -sSL https://api.github.com/repos/Smile-SA/argoos/releases                 \
        -H 'Authorization: token '${ARGOOS_GIT_TOKEN}                               \
        -H 'Accept: application/vnd.github.v3+json'                                 \
        | jq 'to_entries[] | select(.value.tag_name=="'$TAG'") | .value.upload_url' \
        | sed 's/{.*}//'                                                            \
        | sed 's/"//g'
}

# build the release
function build() {
    #go build -tags netgo -ldflags '-X main.VERSION='$TAG
    CGO_ENABLED=0 GOGC=off go build -ldflags '-X main.VERSION='$TAG -a -installsuffix nocgo -o argoos
    strip argoos
    [ $(./argoos -version) == ${TAG} ] || error "Bad version from argoos -version :: "$(./argoos -version)
}


# Upload to github
function upload() {
    url=$1'?name=argoos-linux-x64_64'
    echo "Sending to "$url >&2
    curl -v -SL                                       \
        -H 'Content-Type: application/octet-stream'   \
        -H 'Accept: application/vnd.github.v3+json'   \
        -H 'Authorization: token '${ARGOOS_GIT_TOKEN} \
        --data-binary @argoos $url
}

case $1 in
    build)
        build;
        ;;
    release)
        [ $TAG == "master" ] && error "couldn't release master"
        createRelease;
        ;;
    upload)
        [ $TAG == "master" ] && error "couldn't upload master"
        REL_URL=$(getUploadURL);
        upload $REL_URL;
        ;;
    *)
        echo $(basename $0) "build|release|upload <TAGNAME>"
        exit 1
        ;;
esac

exit 0

