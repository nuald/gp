# Git/p4 helper

A CLI tool to help with Git/p4 interaction.

## Requirements

The tool utilizes Git and P4 command line utilities, please be sure
to install those first.

## Installation

    go get -u github.com/nuald/gp

## Usage

The most common workflow is supported with `clone`, `rebase` and `submit` (`shelve`) commands:

```
NAME:
   gp - Git/p4 helper

USAGE:
   gp [global options] command [command options] [arguments...]

VERSION:
   0.0.1

COMMANDS:
     clone    Creates a new Git directory from an existing p4 repository
     rebase   Updates the Git repository with recent changes from p4
     submit   Submits changes back to the p4 repository
     shelve   Shelves changes back to the p4 repository
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

Please refer to [Git-p4](https://git-scm.com/docs/git-p4) documentation for the details.

The sample workflow:

    $ gp clone //depot/stream local_copy
    $ cd local_copy
    $ ... do the changes ...
    $ git commit -am"message"
    $ gp shelve

`gp clone` creates a new Git directory from an existing p4 repository specified by the depot and the project (or the stream) paths:

    gp clone //depot/project
    gp clone //depot/stream destination

To reproduce the entire p4 history in Git, please use the @all modifier on the depot path:

    gp clone //depot/project@all

The environment and credentials are saved in the global Git config, please
use `--reset-credentials` to clear the config values:

    gp clone //depot/project --reset-credentials
