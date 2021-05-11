# sincap-common

common libs, utils

- [chi](https://github.com/go-chi/chi) for routing.
- [structs](https://github.com/fatih/structs) for reflection.
- [zap](https://github.com/uber-go/zap) for logging.
- [melody](https://github.com/olahol/melody) for websockets.
- [gorm](https://github.com/jinzhu/gorm) for persisting.
- [testify](https://github.com/stretchr/testify) for asserting.

# Test

```bash
go get -u github.com/gophertown/looper
looper
```


## Query API

Multi level searches only works with SingularTableNames for PolymorphicModel and for equals

### Field selection
Give the API consumer the ability to choose returned fields. This will also reduce the network traffic and speed up the usage of the API.

```
GET /cars?_fields=manufacturer,model,id,color
```
### Preload selection
Give the API consumer the ability to choose preloaded (eager) relations. This will also reduce the network traffic and speed up the usage of the API.

```
GET /cars?_preloads=manufacturer,model,id,color
```

### Paging

```
GET /cars?_offset=10&_limit=5
```
* Add  _offset and _limit (an X-Total-Count header is included in the response).

* To send the total entries back to the user use the custom HTTP header: X-Total-Count.

* Content-Range offset – limit / count.

	* offset: Index of the first element returned by the request.

	* limit: Index of the last element returned by the request.

	* count: Total number of elements in the collection.

* Accept-Range resource max.



### Sorting

* Allow ascending and descending sorting over multiple fields.
* Use sort with underscore as `_sort`.
* In code, descending describe as ` - `, ascending describe as ` + `.

```GET /cars?_sort=-manufactorer,+model```

### Operators
* Add `_filter` query parameter and continue with field names,operations and values separated by `,`.
* Pattern `_filter=<fieldname><operation><value>`.
* Supported operations.
	* `=` equal
	* `!=` not equal
	* `<` less
	* `<=` less or equals
	* `>` greater
	* `>=` greater or equals
	* `~=` like
	* `|=` in (values must be separated with `|`
	* `*=` in alternative (values must be separated with `*`
* NULL/mull/nil id reserved word. For ex. Name=NULL or Name!=NULL becomes IS NULL or IS NOT NULL
	
```GET http://127.0.0.1:8080/app/users?_filter=name=seray,active=true```

### Full-text search

* Add `_q`.

```GET /cars?_q=nissan```
