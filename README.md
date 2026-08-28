# go-games-site

The rewrite of my games site into go. Front end will be in react.

## Structure

`ui/` - The react ui contents, this gets built into the binary for the site.

`cmd/server` - The code that links the api and webserver together and contains the main executable.

`internal/web` - The go webserver logic.

`internal/api` - The api that actually runs all the game logic.
