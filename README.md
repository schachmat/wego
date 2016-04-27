**wego** is a weather client for the terminal.

##Features

![Screenshots](https://cloud.githubusercontent.com/assets/8785474/14843914/03683c26-0c11-11e6-92ae-62b0405cca95.png)

* show forecast for 1 to 7 days
* nice ASCII art icons
* displayed info (metric or imperial units):
  * temperature
  * windspeed and direction
  * viewing distance
  * precipitation amount and probability
* ssl, so the NSA has a harder time learning where you live or plan to go
* multi language support
* config file for default location which can be overridden by commandline
* Automatic config management with [ingo](https://github.com/schachmat/ingo)

##Dependencies

* A [working](https://golang.org/doc/install#testing) [Go](https://golang.org/)
  [1.5](https://golang.org/doc/go1.5) environment (You can use
  [goenv](https://github.com/pwoolcoc/goenv) if your distribution does not
  support Go 1.5 yet)
* utf-8 terminal with 256 colors
* A sane monospaced font containing all the required runes (I use `dejavu sans
  mono`)
* An API key for the backend (see Setup below)

##Installation

To install or update the wego binary into your `$GOPATH` as usual, run:
```shell
go get -u github.com/aadithyakv/wego
```

##Setup

0. Run `wego` once. You will get an error message, but the `.wegorc` config file
   will be generated in your `$HOME` directory (it will be hidden in some file
   managers due to the filename starting with a dot).
0. __With a [forecast.io](http://forecast.io/) account__ (new default)
    * Create your account on https://developer.forecast.io/register
    * Update the following `.wegorc` config variables to fit your needs:
    ```
      backend=forecast.io
      location=40.748,-73.985
      forecast-api-key=YOUR_FORECAST.IO_API_KEY_HERE
    ```
0. __With a [Worldweatheronline](http://www.worldweatheronline.com/) account__
    * Worldweatheronline no longer gives out free API keys. [#83](https://github.com/schachmat/wego/issues/83)
    * Update the following `.wegorc` config variables to fit your needs:
    ```
      backend=worldweatheronline.com
      location=New York
      wwo-api-key=YOUR_WORLDWEATHERONLINE_API_KEY_HERE
    ```
0. You may want to adjust other preferences like `days`, `units` and `…-lang` as
   well. Save the file.
0. Run `wego` once again and you should get the weather forecast for the current
   and next few days for your chosen location.
0. If you're visiting someone in e.g. London over the weekend, just run `wego 4
   London` or `wego London 4` (the ordering of arguments makes no difference) to
   get the forecast for the current and the next 3 days. Unfortunately that does
   not currently work with the forecast.io backend, as it only supports
   latitude,longitude location specification.

You can set the `$WEGORC` environment variable to override the default config
file location.

##Todo

* more [backends and frontends](https://github.com/schachmat/wego/wiki/How-to-write-a-new-backend-or-frontend)
* resolve ALL the [issues](https://github.com/schachmat/wego/issues)
* don't forget the [TODOs in the code](https://github.com/schachmat/wego/search?q=TODO&type=Code)

##License - ISC

Copyright (c) 2014-2016,  <teichm@in.tum.de>

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
