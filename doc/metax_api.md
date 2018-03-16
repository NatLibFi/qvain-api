# Metax Api
===========

## Preliminary

This document describes briefly which Metax API calls Qvain depends on. Any change in these responses requires an accompanying change in Qvain.

## Dataset records

### `/api/dataset`
------------------

_query for a user's dataset records_

#### Methods

> HEAD:
> GET:
	_return all datasets for the user with owner_id `<uuid>`_
	
	request params:
		ownerid=<uuid>
	request headers:
		modified-since
	response headers:
		Content-Type: application/json
		X-Count: integer number of datasets that will be served by the response
	response format:
		JSON "stream" format, list of objects: '[' + n*<{dataset},> + ']'
	notes:
		- Content-Length not necessary
		- Content generation for `HEAD` request not necessary: could be a simple SQL count(*)
		- API consumed by Go which is a static language, so mind the data types:
			- a `string` can't also be `null` without writing extensive and slow introspective data type checks
			- missing object fields are ok


## File API

### `/rest/directories/files?include_parent=true&project=<project>&path=<path>`

_query for files and directories at path_

> GET:
	_return all files and directories for project <project> path <path>_

	request params:
		include_parent=true
		project=<project>
		path=<unix path>
	request headers:
		_(none)_
	response headers:
		Content-Type: application/json
	response format:
		JSON, with:
		- directories[].directory_name (and other properties)
		- files[].file_name (and other properties)
		- .id
		- .directory_name
		- .directory_path
	notes:
		- API client is javascript, Vue with Axios library
