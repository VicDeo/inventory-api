# Project Title

Inventory Management system.
Written as a "API Development and Documentation using Go" course project.

## Getting Started

This application is designed to manage inventory via RESTful API.
It supports the following endpoints:
* `GET /inventory/{id}` - Get a single item by id
* `GET /inventory/` - Get all items
* `POST /inventory/` - Create an item
* `PATCH /inventory/{id}` - Update one or more item properties by id
* `PUT /inventory/{id}` - Update all item properties by id
* `DELETE /inventory/{id}` - Delete item by id

For more detailed build-in documentation on API start the app and browse to `/swagger/index.html`.

The app has a built-in rate limiting using the Token Bucket algorithm.
Number of requests per second is configurable via `APP_RPS_LIMIT` env variable.

### Dependencies

To run this project you need the following programs installed on your system.

```
make v4.4.1
docker v27.5.1
go v1.22.12
swag v1.16.4
```

Versions are listed for the reference only. But we appreciate if your environment will have versions no lower than ours.

### Installation

1. Clone this repo
2. cd to the cloned repo directory
3. Update variables in .env file 
4. Execute `make up-dev` to start a local Postgres server
5. Execute `make build-docs` to generate documentation
6. Execute `make run` to start the local app 
7. If you have a great passion towards testing you may execute `make test` as well

```
git clone https://github.com/VicDeo/inventory-api.git
cd inventory-api
make up-dev
make build-docs
make run
```

Now you're ready to browse the app at the location specified in the `APP_URL` var of the .env file

## Testing

Run `make up-dev` if you haven't been run it yet to start a database
Run `make run` to start the app
Run `make test` to start testing

### Break Down Tests

Unfortunately our developers haven't provided any tests yet but we swear to write some by the release of the version 100.0.0

```
Just imagine we have examples here
```

## Built With

* **Golang** - My favorite programming language

## License

MIT. Take it, Like it, Share it, Break it. 