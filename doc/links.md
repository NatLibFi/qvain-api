# Project links

This document lists links to technical resources related to Qvain, the National Library of Finland's contribution to the ATT/FAIR Data project; consider it a set of bookmarks for developers.

Note that some of the linked services might require authentication.


## Qvain

_links to Qvain related resources_

- Qvain backend:
  - [Github](https://github.com/CSCfi/qvain-api) repository
  - [Github issues](https://github.com/CSCfi/qvain-api/issues) – report bugs here
  - [Jira bug tracker](https://jira.eduuni.fi/projects/CSCQVAIN) at CSC – for planning and service integration
  - Documentation:
    - [Rest API](./api.md)
    - [Metax API](./metax_api.md) dependencies
    - [Go package dependencies](./go_dependencies.md) [auto-generated]
- Qvain frontend:
  - [Github](https://github.com/CSCfi/qvain-js) repository
- [General overview](https://www.kiwi.fi/display/ATT/Technical+overview) at the National Library of Finland's [ATT/Tajua project wiki](https://www.kiwi.fi/pages/viewpage.action?pageId=53839580)
- [Roadmap](https://github.com/CSCfi/qvain-js/blob/next/doc/roadmap.md) including the what, why and how of Qvain


## Metax

_links to CSC's Metax related resources_

- [Github repository](https://github.com/CSCfi/metax-api)
- [Jira bug tracker](https://jira.eduuni.fi/projects/CSCMETAX)
- Eduuni wiki:
  - [Development](https://wiki.eduuni.fi/display/CSCMETAX/Development)
  - [API documentation](https://wiki.eduuni.fi/display/CSCMETAX/API+documentation)
  - [Database](https://wiki.eduuni.fi/display/CSCMETAX/Database+documentation)
  - [Reference data](https://wiki.eduuni.fi/display/CSCMETAX/Reference+Data)
- Flowdock: [development](https://www.flowdock.com/app/tiptop/metax-kehitys) and [general](https://www.flowdock.com/app/tiptop/metax) chat
- Metax API (test instance):
  - [rest api](https://metax-test.csc.fi/rest/)
  - [dataset api](https://metax-test.csc.fi/rest/datasets/)
  - [dataset stream](https://metax-test.csc.fi/rest/datasets/?no_pagination=true&owner_id=055ea531a6cac569425bed94459266ee&stream=true)
  - [file api, directories](https://metax-test.csc.fi/rest/directories/2)
  - [file api, files](https://metax-test.csc.fi/rest/files/)
  - [file api, path](https://metax-test.csc.fi/rest/directories/files?project=project_x&path=/project_x_FROZEN/Experiment_X&include_parent)
- Metadata schemas:
  - [base](https://github.com/CSCfi/metax-api/tree/test/src/metax_api/api/rest/base/schemas)
  - [api](https://github.com/CSCfi/metax-api/tree/test/src/metax_api/api/rest/base/api_schemas)


## IOW / Tietomallit

_links to metadata schemas_

- [Metax Research Datasets](https://tietomallit.suomi.fi/model/mrd/) – the main metadata schema used in Qvain
- [Metax Research Datasets (IOW)](http://iow.csc.fi/model/mrd/CatalogRecord/)
- [Metax Administrative Contracts](http://iow.csc.fi/model/mad/)
- [Metax Research Data Catalogs](http://iow.csc.fi/model/mdc/Catalog/)
- [Metax Data Storage Metadata](http://iow.csc.fi/model/mfs/)
- [Metax Statistics](http://iow.csc.fi/model/mstat/)
- [Etsin tietomalli (fi)](http://iow.csc.fi/model/etsin/)


## FairData AAI

_links to FairData ID service_

- [Jira planning and status](https://jira.eduuni.fi/projects/CSCFAIRDATAAAI)
- [Meeting notes](https://wiki.eduuni.fi/pages/viewpage.action?pageId=54699141)
- [User management](https://wiki.eduuni.fi/pages/viewpage.action?pageId=44569948)

## Other
- [Integration tests](https://github.com/CSCfi/FAIRDATAtest) for FairData services
- [Tutkimus-PAS architectural overview](https://wiki.eduuni.fi/display/att/Tutkimus-PAS+arkkitehtuurikuvauksia)
- [Fairdata style guide](https://wiki.eduuni.fi/display/att/Tutkimus-PAS%3A+visuaaliset+perusasiat)
