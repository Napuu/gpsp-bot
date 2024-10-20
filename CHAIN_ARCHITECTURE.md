## Handlers
* Service-specific message handlers
  * TelegramMessageHandler
  * DiscordMessageHandler
* GenericMessageHandler - parses action
* URLParser - extracts urls from message (if action requires that?)
* VideoDownloader - downloads the video (and starts status daemon thingy on its own channel)
* TuplillaHandler - extra
