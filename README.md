# stock-archiver

Add the tickers for which you'd like csv logging as program args:

`--s AAPL --s GOOG --s MSFT --i 5`

Would log data for Apple, Google, Microsoft every 5 seconds.

Currently supports logging timestamp, regular_market_price, bid_price, bid_volume, ask_price, ask_volume.

Completely powered by Yahoo Finance API.
