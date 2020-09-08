# ancient-store-s3
> Store your core-geth ancient block data on S3.

This program is a persistence manager for core-geth "ancient" block data using Amazon S3,
and can be used as a replacement for the core-geth (and go-ethereum) default of persisting ancient data in local flat, compressed files.

## Requirements

- [core-geth](https://github.com/etclabscore/core-geth)

## Setup Environment Variables

You'll need to use AWS's profile configuration strategies, eg. an `~/.aws/config` file. Once you have a profile established for use,
you can set the corresponding environment variable.

```sh
export AWS_REGION=us-east-1 
export AWS_PROFILE=developers-s3
```

## Install

```sh
> git clone https://github.com/etclabscore/ancient-store-s3.git
> cd ancient-store-s3
> go build .
> ./ancient-store-s3 --help
```

## Run ancient server
```
./ancient-store-s3 --ipcpath /path/to/ancient.ipc
```
### Run core-geth with ancient RPC path
```sh
geth --ancient.rpc /path/to/ancient.ipc
```


