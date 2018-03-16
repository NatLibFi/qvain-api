# Qvain Api
===========

## Preliminary

### About this document

This document is maintained in tandem with the backend code of Qvain and should be an accurate reflection of the current behaviour of the application.


### Mime format and character set

The mime format is `application/json`, the default for JSON. All in- and output should be in utf-8 encoded unicode.


### Authentication and authorization

All requests need to be authenticated. The server looks for a valid bearer token in the `Authorization` header of the HTTP request (see [rfc 6750](https://tools.ietf.org/html/rfc6750#section-2.1)).

Requests without valid authentication will return `401 Unauthorized`.

Requests with valid authentication but on datasets for which the user does not have permissions will return `403 Forbidden`.


### Errors

Errors return an HTTP status code in the `4xx` or `5xx` range (if possible) with a JSON payload.

```json
{
	"error": {
		"code": 400,
		"message": "this is an error message"
	}
}
```


## API endpoints

### `/api/dataset`
------------------

_operations on the Qvain record set_

#### Methods

> GET:
		_list Qvain records of datasets_

		params: (filter) creator=<uuid>, owner=<uuid>
		returns: 200
		status: not implemented

> POST:
		_create a new Qvain dataset record_

		returns: 201 + redir
		status: not implemented


### `/api/dataset/<uuid>`
-------------------------

_operations on a Qvain record_
	
#### Notes

These are operations on the Qvain-specific record of a dataset, not the actual dataset blob itself.

#### Methods

>	GET
		_retrieves a full record_

		returns: 200
		status: not implemented (needed?)

>	PUT
		_saves (overwrites) a record_

		returns: 200
		status: not implemented (needed?)

>	PATCH
		_changes given top-level fields_

		returns: 200
		status: implemented

>	DELETE
		_deletes a dataset_

		returns: 200
		status: not implemented (needed?)


### `/api/dataset/:uuid/[keypath]`
----------------------------------

_operations on the dataset in a Qvain record_

#### Notes

Not all datasets allow all of their contents to be gotten or set.
For instance for Metax records only specific paths can be edited through the API (e.g. `/metadata`, `/files` and `/contracts`).

#### Methods

>	GET
		_retrieves all or part of a dataset record_

		returns: 200
		status: not implemented

>	PUT
		_sets all or part of a dataset record_

		returns: 200
		status: not implemented



# Record [/api/record]

## Retrieve All Posts [GET]
+ Response 200 (application/json)
    + Attributes (array[Blog Post])

meh.
