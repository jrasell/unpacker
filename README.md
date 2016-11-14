# Unpacker

Unpacker is a tool designed to tidy up after Packer which can often leave orphaned resources.

It currently only supports cleaning up after the Amazon EC2 Packer builder.

## Installation

#### Binaries


#### Docker

```bash
$ docker pull jrasell/unpacker
```

#### Go

```bash
$ go get github.com/jrasell/unpacker
```

## Usage

- **`--dryrun`** Performs a dry-run, only printing what resources would be removed.
- **`--region`** The AWS region to connect to a look for Packer resources.
- **`--tag_key`** The AWS EC2 tag key which presents your Packer build instances.
- **`--tag_value`** The AWS EC2 tag value which presents your Packer build instances.

## Contributing

Any contributions are much appreciated. If you would like to contribute please open a pull-request.


[![Analytics](https://ga-beacon.appspot.com/UA-87275628-1/unpacker/README.md)](https://github.com/jrasell/unpacker)