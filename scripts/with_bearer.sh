#!/bin/sh
#
# -wvh- run curl with bearer token
#

TOKENGEN="go run ${HOME}/Code/GoPath/src/wvh/makejwt/main.go -token -aud $(hostname --fqdn)"
TOKEN=""

if [ -z "$1" ]; then
	echo usage: "$0 <url>"
	exit 1
fi

if [ -f ~/Code/GoPath/src/wvh/makejwt/main.go ]; then
	TOKEN=$(${TOKENGEN})
fi

if [ -z ${TOKEN} ]; then
	echo $0: "warning: can't generate new token; using old token (probably expired)"
	TOKEN='eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJzZXJ2aWNlLmV4YW1wbGUuY29tIiwiZXhwIjoxNTE4NTI0NDI2LCJpYXQiOjE1MTg1MjQzMzYsInByaXZhdGVDbGFpbUtleSI6IkhlbGxvLCBXb3JsZCEiLCJzdWIiOiIwNTY1MTZmZmEzM2YwZjJhMmZhMmMzMzRhMTNlODUxNiJ9.kkax1nWcxec4fnWWuH40wj7RoZKUR2uHks-Vq_Gir3U'
fi


curl -H "Accept: application/json" -H "Authorization: Bearer ${TOKEN}" "${@}"
