#wego

wego is a weather client for the terminal.

##Features

* show forecast for 1 to 5 days
* nice ASCII art icons
* displayed info (metric or imperial units):
  * temperature
  * windspeed and direction
  * viewing distance
  * precipitation amount and probability
* ssl, so the NSA has a harder time learning where you live or plan to go
* config file for default location which can be overridden by commandline
* multi language support

![Screenshots](http://schachmat.github.io/wego/wego.gif)

##Dependencies

* Working Go environment
* utf-8 terminal with 256 colors (I use the dejavu font)
* A worldweatheronline.com API key (see Setup below)

##Setup

1. To Install the wego binary into your `$GOPATH` as usual, run:
   `go get -u github.com/schachmat/wego`
2. Run `wego` once. You will get an error message, but the config file will be
   generated for you as well.
3. If you don't have the necessary API key yet, you can [register
   here](https://developer.worldweatheronline.com/auth/register) with your
   github.com account. Your github.com account needs a public email address, but
   you can choose a bogus one.
4. Copy your API key into the `.wegorc` file in your `$HOME` folder, change
   the city to your preference, and choose if you want to use metric or imperial
   units. Save the file.
5. Run `wego` once again and you should get the weather forecast for the current
   and next 2 days.
6. If you're visiting someone in e.g. London over the weekend, just run
   `wego 4 London` or `wego London 4` (the ordering of arguments makes no
   difference) to get the forecast for the current and the next 3 days.

* You can also set the `$WEGORC` environment variable to override the default
  location.

##Todo

* store values affected by metric/imperial conversion as float32 internally
* implement SI units
* use objx instead of custom marshall function to localize the output (and
  handle the whole json response in general)

##License

Copyright (c) 2015,  <teichm@in.tum.de>

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.
