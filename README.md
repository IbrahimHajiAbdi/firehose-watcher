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
### Linux
```bash
go build -o fw
```

### Windows
```bash
go build -o fw.exe
```

## Usage
``fw`` will download the media attached to each post and metadata about the post itself to a given directory. 

The files will be named in this format: ``{rkey}_{handle}_{text}``.

### Example

#### Linux/Unix
```bash
./fw --handle bsky.app path/to/directory/
```

#### Windows
```bash
./fw.exe --handle bsky.app path/to/directory/
```

## Options
- ``--handle``
  - The handle of the account you want to subscribe to. **Required**