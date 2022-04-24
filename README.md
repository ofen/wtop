`wtop` looks for wallets ([ETH](https://ethereum.org)) whose balance has changed the most (in positive or negative direction) between selected blocks. Requires [getblock.io](https://getblock.io/) API token (see https://getblock.io/docs/get-started/auth-with-api-key/).

## Usage
Search between `head-30` and `head` blocks
```sh
wtop -n 30 -t 2d270fd9-ffe8-46fe-9eba-0e4662a417ea
```

Search between `14648000` and `14648260` blocks
```sh
wtop -n 260 -t 2d270fd9-ffe8-46fe-9eba-0e4662a417ea 14648260
```

## Build
```sh
go build .
```