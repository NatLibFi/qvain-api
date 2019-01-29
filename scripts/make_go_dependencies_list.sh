#!/bin/sh
#
# -wvh- automatically generate a list of Go package dependencies for documentation purposes
#

OUTFILE="doc/go_dependencies.md"
BASEDIR=$(dirname ${0})/..
CODEBLOCK='```'
PREFIXCMD='sed s/^/\t/'

set -e

#if [ ! -e ${0##*/} ]; then
#	echo "$0: error: please run this script from its containing directory"
#	exit 1
#fi

if [ ! -e ${BASEDIR}/doc ]; then
	echo "$0: error: can't find doc/ folder in this script's parent directory"
	exit 1
fi

{
	cd ${BASEDIR}
	# base of own import path; find longest common prefix
	SELF=$(go list ./... | sed -e 'N;s/^\(.*\).*\n\1.*$/\1\n\1/;s,/$,,;D')
	cat <<-__EOF__
	# Go package dependencies
	-------------------------

	This auto-generated document lists all imports and indirect package dependencies, excluding the packages that come with Go itself as part of the standard library.

	### Imports (directly imported packages)
	
	$(go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' $(go list -f '{{range .Imports}}{{.}} {{end}}' ./cmd/...) | grep -vE "^${SELF}/|/internal$" | sort | ${PREFIXCMD} || echo "error generating list")

	### Dependencies (packages required by imported packages)
	
	$(go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' $(go list -f '{{range .Deps}}{{.}} {{end}}' ./cmd/...) | grep -vE "^${SELF}/|/internal$" | sort | ${PREFIXCMD} || echo "error generating list")

	-- 
	generated on $(date "+%Y-%m-%d") at commit $(git describe --always || echo "_unknown_") by ${0##*/}
	__EOF__
} > ./${OUTFILE}

