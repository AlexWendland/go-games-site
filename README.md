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

## Environment variables

Environment variables are read at start time and processed by `/internal/config/environment.go`.

| Variable name | Type | Default | Purpose |
| ------------- | ---- | ------- | ------- |
| GAMES_PRODUCTION | bool | false   | If this is running in production and should turn off any CORS hokey pokey |
| GAMES_PORT    | int  | 8000    | The port to host the server on |
| GAMES_BASE_URL | string | http://localhost:8080 | The base url the server is running on |
| GAMES_ALLOWED_ORIGINS| []string | http://localhost:5173 | The origins allowed for CORS - mainly used if you want to run the server UI in auto reload |
| GAMES_DATABASE_FILE | string | games.db | The location of the game database file |

## Developer notes

When setting up the repo you will want to run:

```
git update-index --assume-unchanged ui/dist/.gitkeep
```

This will leave the .gitkeep file untracked as make will delete it upon clean.
