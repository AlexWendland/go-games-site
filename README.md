# go-games-site

The rewrite of my games site into go. Front end will be in react.

## Structure

`ui/` - The react ui contents, this gets built into the binary for the site.

`cmd/server` - The code that links the api and webserver together and contains the main executable.

`internal/web` - The go webserver logic.

`internal/api` - The api that actually runs all the game logic.

`internal/db` - The code that relates to making changes and updating the database.
This is mostly auto generated using [sqlc](https://docs.sqlc.dev/en/latest/) and controlled by the migrations directory (powered by [goose](https://pressly.github.io/goose/)) and the queries directory containing the sql that it runs.

## Database

The backend is stateless and instead saves everything down to a sqlite database.
