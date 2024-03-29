
# confman
[![Generic badge](https://img.shields.io/badge/License-MIT-blue.svg)](https://github.com/capsic/confman/blob/main/LICENSE) [![Generic badge](https://img.shields.io/badge/Made_with-Go-green.svg)](https://go.dev)

Simple configuration manager implemented as REST-ful microservice, written in [Go](https://go.dev/).

Useful if you have a project that consists of multiple applications/modules. `confman` provides a secure and easy way to maintain centralized configuration file that can be conveniently shared and accessed by your applications/modules across the whole project.

## Installation

Install all dependencies.
Skip any dependencies that you already have in your Go installation.

```bash
go get -u github.com/gorilla/mux
go get -u github.com/tidwall/gjson
go get -u golang.org/x/term
```

Clone and build this repository.

```bash
git clone https://github.com/capsic/confman
go build .
```

## Configuration
**`/config.json`** - `confman` configuration

```json
{
    "port": 7777,
    "configurationFile": "configuration.conf",
    "encrypt": true,
    "ipWhitelist": [
        "127.0.0.1",
        "[::1]"
    ],
    "ssl": true,
    "certFile": "fullchain.pem",
    "keyFile": "privkey.pem"
}
```
1. `port` - The microservice will be bound to this port.
2. `configurationFile` - The configuration file name (inside `data` directory) that will be served by `confman`.
3. `encrypt` - Configuration file may contain sensitive informations (eg. database password, private IP address, email info, etc.), you might want to enable this option so `confman` will encrypt the original configuration file once you start the service. You will be prompted to enter encryption passphrase when you start the service. 
4. `ipWhitelist` - Additional security measure to limit access by remote IP address. Empty array means no IP checking.
5. `ssl` - Enable/disable SSL support on the service. If enabled you must place your *certificate file* (eg. cert.pem) and *private key file* (eg. key.pem) in the `cert` directory.
6. `certFile` - SSL certificate file name, ignore if `ssl` is disabled.
7. `keyFile` - SSL private key file name, ignore if `ssl` is disabled.


**`/data/configuration.conf`** - the actual configuration file that will be served by `confman`, put whatever you need in here. Should be in JSON format, elements can be nested arbitrarily.
```json
...
{
    "mysql": {
        "host": "127.0.0.1",
        "port": 3306,
        "user": "mysqluser",
        "password": "mysqlpassword"
    },
    "rabbitmq": {
        "host": "127.0.0.1",
        "port": 5672,
        "vhost": "/",
        "credentials": {
            "user": "rabbituser",
            "password": "rabbitpassword"
        }
    },
    "someArray": [
        {"id": 1, "name": "John"},
        {"id": 2, "name": "Doe"},
        {"id": 3, "name": "Jane", "data": ["a", "b", "c", 1, 2, 3]}
    ]
}
...
```

## Usage

Just execute the Go binary that you got at build.
By default `CONFMANHOME` is set to `/opt/capsic/confman`, if you're not running from this directory you need to specify the absolute path you're running this service from. This can be done either by:

1. Manually set `CONFMANHOME` env. variable to the directory of your `confman` build.
2. Or, specify it as CLI argument to the command (example #2).
3. Or, change the `DEFAULTHOME` constant in `main.go` then rebuild the binary.

```bash
./confman
./confman /Users/myuser/confman
```

Making a REST request (example):
```
http://127.0.0.1:7777/get?key=mysql
>> {"host":"127.0.0.1","password":"mysqlpassword","port":3306,"user":"mysqluser"}

http://127.0.0.1:7777/get?key=mysql.host
>> "127.0.0.1"

http://127.0.0.1:7777/get?key=rabbitmq.port
>> 5672

http://127.0.0.1:7777/get?key=rabbitmq.credentials.password
>> "rabbitpassword"

https://127.0.0.1:7777/get?key=someArray
>> [{"id":1,"name":"John"},{"id":2,"name":"Doe"},{"id":3,"uname":"Jane","username":"janedoe"}]

https://127.0.0.1:7777/get?key=someArray.2.data
>> ["a","b","c",1,2,3]

https://127.0.0.1:7777/get?key=someArray.2.data.0
>> "a"
```

## Credits
[github.com/gorilla/mux](https://github.com/gorilla/mux)

[github.com/tidwall/gjson](https://github.com/tidwall/gjson)

[golang.org/x/term](https://pkg.go.dev/golang.org/x/term)


## License
[MIT](https://github.com/capsic/confman/blob/main/LICENSE)
