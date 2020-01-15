# link2epub [![Go Report Card](https://goreportcard.com/badge/github.com/arnaucube/link2epub)](https://goreportcard.com/report/github.com/arnaucube/link2epub)
Very simple tool to download articles and convert it to `.epub`/`.mobi` files.

It gets the text content, simplifies its html, downloads the images, and builds the `.epub`/`.mobi` file.

## Download
- Binary can be:
	- downloaded from [releases section](https://github.com/arnaucube/link2epub/releases)
	- compiled with `go build`

## Usage
Needs [calibre](https://calibre-ebook.com/) in order to convert to `.epub` and `.mobi`.

Putting the binary in the `~/bin` directory will be more comfortable.

```bash
link2epub -l https://link.com/to-the-article

// optionally add extension (by default .mobi)
link2epub -l https://link.com/to-the-article -type mobi
link2epub -l https://link.com/to-the-article -type epub

// see help for all the available flags
link2epub --help
```

Thanks to [@dhole](https://github.com/dhole) for the advisment.

