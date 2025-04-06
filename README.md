# General Purpose S-Posting Bot

Add useful and not so useful commands to group chats. Reusing business logic so same bot can run on Telegram and Discord.

## Available commands

Commands must be explicitely enabled via an environment variable. E.g., `ENABLED_FEATURES=ping;dl`.

### /dl \<link>

Downloads video in the link and sends it as a video while deleting the original message if everything succeeds.

Use `PROXY_URLS` environment variable to provide list of available SOCKS5-proxies to circumvent IP-range restrictions. Videos are stored at `YTDLP_TMP_DIR` which is cleared periodically.

This command requires [yt-dlp](https://github.com/yt-dlp/yt-dlp) to be available at `PATH`. For downloads to keep working, you should run `yt-dlp -U` so it is running the latest available release.

### /euribor

Values are fetched from Suomen Pankki dashboards with Playwright and cached at SQLite db located at `DATABASE_FILE`. Playwright used because the "official" dashboards available through APIs are not updated as quickly as this one dashboard.
```
Euribor-korot 04.04.
12 kk: 2.235 %
6 kk: 2.259 %
3 kk: 2.323 %
```
### /tuplilla \<jotain>

Dubz. Throws two dice responds in a proper way.
```
/tuplilla tehdään asia X
dubz:    Tuplat tuli 😎, tehdään asia X
no dubz: Ei tuplia 😿, ei tehdä asiaa X
```
`MISTRAL_TOKEN` required for this one because we try respond with grammatically correct negation. Sometimes works, sometimes not.


## Sample configs

### Telegram with a lot of features enabled
```
PROXY_URLS="localhost:1235;localhost:1234" \
  YTDLP_TMP_DIR=/tmp/yt-dlp \
  EURIBOR_GRAPH_DIR=/tmp/euribor \
  DATABASE_FILE=/opt/euribor.db \
  MISTRAL_TOKEN=<mistral token> \
  TELEGRAM_TOKEN=<telegram token> \
  ENABLED_FEATURES="ping;dl;euribor;tuplilla" \
  go run gpsp-bot.go telegram
```

### Discord with not so much features enabled
```
DISCORD_TOKEN=<telegram token> \
  ENABLED_FEATURES=ping \
  go run gpsp-bot.go discord
```

### Running built container with yt-dlp mounted from host
Container includes yt-dlp at build-time, but see /dl -section why it is useful to mount yt-dlp from host.
```
podman built -t gpsp-bot .
podman run -v /usr/bin/yt-dlp:/usr/bin/yt-dlp:z \
  -e ENABLED_FEATURES="ping;dl" \
  -e TELEGRAM_TOKEN=<token> gpsp-bot telegram
```