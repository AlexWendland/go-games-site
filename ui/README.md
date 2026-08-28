# Frontend

This front end is built into the go binary - this is linked up using the embed.go file (only the `/dist` directory).
This front end uses react with vite.

## How this works

It all starts from `index.html` this imports in `src/main.tsx` which starts off all the react stuff.

We use tailwind for all the css stuff.

## Images

For images that need a 'stable' url for use within the html files (such as index.html) use the `public/` folder.
For images that can be optimised via vite such as game components put them in the `src/assets` folder.
