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

## Troubleshooting

1045 Access Denied error on MySql :

- Be sure username and password is 'spin'
- Be sure mysql service is up and running.

## Query API

Multi level searches only works with SingularTableNames for PolymorphicModel and for equals
