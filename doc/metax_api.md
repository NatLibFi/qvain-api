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


## Sync process

Description of the synchronisation logic between Qvain and Metax.

- Qvain connects to Metax dataset endpoint:
	- query parameters:
		- no_pagination=true
		- stream=true
		- owner_id=<uuid>
	- headers:
		- `If-Modified-Since`
- Metax responds in JSON format with an array of dataset objects and `X-Count` header set to the number of objects that will be returned;
	- if `owner_id` was included in the request, only records for the given Qvain owner uuid will be returned;
	- if `owner_id` was not included, all records relevant to Qvain will be returned;
	- if `If-Modified-Since` was provided in the request's headers, only records with a modification time after the one provided will be returned;
	- the response shall not included any other fields such as paging because it will be parsed as a stream.
- Qvain parses the stream per object, looking for the Qvain metadata block `dataset.editor` with the following fields:
	- `dataset.editor.dataset_id`
	- `dataset.editor.identifier`
	- `dataset.editor.creator_id`
	- `dataset.editor.owner_id`
- Qvain validates the metadata block:
	- if there is no `dataset.editor object`, Qvain skips the record;
	- if there is no `dataset.editor.identifier` object or it is not the literal string "qvain", Qvain skips the record;
	- if there is a `dataset.editor.dataset_id` value and it parses as a UUID, Qvain overwrites the local record;
	- if there is no `dataset.editor.dataset_id` value, but `dataset.editor.owner_id` is set, Qvain will create a new record with the given owner and the current date;
