# composegen
composegen is a generator of docker-compose service.

## Installation
```
go get -u github.com/orisano/composegen
```

## Commands
### db
composegen db creates docker-compose service from URL syntax connection string.

#### postgres
```
$ composegen db -url postgres://booktest:booktest@localhost/booktest
version: '3'
services:
  postgres:
    image: postgres:latest
    environment:
      POSTGRES_DB: booktest
      POSTGRES_PASSWORD: booktest
      POSTGRES_USER: booktest
    ports:
    - 5432:5432
```

#### mysql
```
$ composegen db -url mysql://booktest:booktest@localhost/booktest
version: '3'
services:
  mysql:
    image: mysql:latest
    command: --default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4
      --collation-server=utf8mb4_unicode_ci
    environment:
      MYSQL_ALLOW_EMPTY_PASSWORD: "yes"
      MYSQL_DATABASE: booktest
      MYSQL_PASSWORD: booktest
      MYSQL_USER: booktest
    ports:
    - 3306:3306
```

#### redis
```
$ composegen db -url redis://booktest:booktest@localhost/booktest
version: '3'
services:
  redis:
    image: redis:latest
    command: --requirepass "booktest"
    ports:
    - 6379:6379
```

## Author
Nao YONASHIRO (@orisano)

## License
MIT
