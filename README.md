# BetterWSC
Wii Sports Club server for Samtendo Network (Credits to @Whatsauser)

## Compiling

### Setup
Install [Go](https://go.dev/doc/install) and [git](https://git-scm.com/downloads), then clone and enter the repository

```bash
$ git clone https://github.com/SamtendoNetwork/BetterWSC
$ cd BetterWSC
```

### Compiling using `go`
To compile using Go, `go get` the required modules and then `go build` to your desired location. You may also want to tidy the go modules, though this is optional

```bash
$ go get -u
$ go mod tidy

$ go build -o build/wii-sports-club
```

The server is now built to `build/wii-sports-club`

### Compiling using `make`
Compiling using `make` will read the local `.git` directory to create a dynamic authentication server build string, based on your repositories remote origin and current commit

Install `make` either through your systems package manager or the [official download](https://www.gnu.org/software/make/). We provide a `default` rule which compiles [using `go`](#compiling-using-go)

To build using `go`

```bash
$ make
```

The server is now built to `build/wii-sports-club`

## Configuration
All configuration options are handled via environment variables

`.env` files are supported

| Name                                | Description                                    | Required                            |
| ----------------------------------- | ---------------------------------------------- | ----------------------------------- |
| `SN_WSC_POSTGRES_URI`               | Fully qualified URI to your Postgres server    | Yes                                 |
| `SN_WSC_AUTHENTICATION_SERVER_PORT` | Port for the authentication server             | Yes                                 |
| `SN_WSC_SECURE_SERVER_HOST`         | Host name for the secure server                | Yes                                 |
| `SN_WSC_SECURE_SERVER_PORT`         | Port for the secure server                     | Yes                                 |
| `SN_WSC_ACCOUNT_GRPC_HOST`          | Host name for your account server gRPC service | Yes                                 |
| `SN_WSC_ACCOUNT_GRPC_PORT`          | Port for your account server gRPC service      | Yes                                 |
| `SN_WSC_ACCOUNT_GRPC_API_KEY`       | API key for your account server gRPC service   | No (Assumed to be an open gRPC API) |
| `SN_WSC_FRIENDS_GRPC_HOST`          | Host name for your friends server gRPC service | Yes                                 |
| `SN_WSC_FRIENDS_GRPC_PORT`          | Port for your friends server gRPC service      | Yes                                 |
| `SN_WSC_FRIENDS_GRPC_API_KEY`       | API key for your friends server gRPC service   | No (Assumed to be an open gRPC API) |
