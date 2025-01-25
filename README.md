# Firehose-Watcher

Firehose-Watcher is a CLI made in Go that subscribes to a given account and downloads their likes, reposts and posts as they happen.

## Installation
You will need [Go](https://go.dev/doc/install) installed.

Clone the repository and ``cd`` into:

```bash
git clone git@github.com:IbrahimHajiAbdi/firehose-watcher.git
cd firehose-watcher
```

Then install all the dependencies:
```bash
go mod tidy
```

Then build the program:

```bash
go build -o fw
```

## Usage
``fw`` will download the media attached to each post and metadata about the post itself to a given directory. 

The files will be named in this format: ``{rkey}_{handle}_{text}``.

### Example

```bash
./fw --handle bsky.app --directory media
```

## Options
- ``--handle``
  - The handle of the account you want to subsribe to
- ``--directory``
  - Where you want to save the posts to