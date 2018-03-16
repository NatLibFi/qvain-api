#!/bin/sh
#
# -wvh- automatically generate a list of Go package dependencies for documentation purposes
#

OUTFILE="go_dependencies.md"
BASEDIR=".."
CODEBLOCK='```'
PREFIXCMD='sed s/^/\t/'

set -e

if [ ! -e ${0##*/} ]; then
	echo "$0: error: please run this script from its containing directory"
	exit 1
fi

{
	cd ${BASEDIR}
	cat <<-__EOF__
	# Go package dependencies
	-------------------------

	This auto-generated document lists all imports and indirect package dependencies, excluding the packages that come with Go itself as part of the standard library.

	### Imports (directly imported packages)
	
	$(go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' $(go list -f '{{range .Imports}}{{.}} {{end}}' ./...) | grep -vE '^wvh/|/internal$' | sort | ${PREFIXCMD} || echo "error generating list")

	### Dependencies (packages required by imported packages)
	
	$(go list -f '{{if not .Standard}}{{.ImportPath}}{{end}}' $(go list -f '{{range .Deps}}{{.}} {{end}}' ./...) | grep -vE '^wvh/|/internal$' | sort | ${PREFIXCMD} || echo "error generating list")

	-- 
	generated on $(date "+%Y-%m-%d") at commit $(git describe --always || echo "_unknown_") by ${0##*/}
	__EOF__
} > ./${OUTFILE}

