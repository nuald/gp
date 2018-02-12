# Git/p4 helper

A CLI tool to help with Git, p4 and Swarm integration.

## Requirements

The tool utilizes Git and P4 command line utilities, please be sure
to install those first.

## Installation

    go get -u github.com/nuald/gp

## Usage

The most common workflow is supported with `clone`, `rebase` and `submit` (`shelve`) commands:

```
Git/p4 helper

Usage:
  gp [command]

Available Commands:
  clone       Create a new Git directory from an existing p4 repository
  help        Help about any command
  rebase      Update the Git repository with recent changes from p4
  review      Add #review hashtag and the list of reviewers into the HEAD commit
  shelve      Shelve changes back to the p4 repository
  submit      Submit changes back to the p4 repository

Flags:
  -c, --clear-credentials   clear saved credentials
  -h, --help                help for gp
      --version             version for gp

Use "gp [command] --help" for more information about a command.
```

Please refer to [Git-p4](https://git-scm.com/docs/git-p4) documentation for the details.

The sample workflow:

    $ gp clone //depot/stream local_copy
    $ cd local_copy
    $ ... do the changes ...
    $ git commit -am"message"
    $ gp rebase
    $ gp shelve

### gp clone

`gp clone` creates a new Git directory from an existing p4 repository specified by the depot and the project (or the stream) paths:

    gp clone //depot/project
    gp clone //depot/stream destination

To reproduce the entire p4 history in Git, please use the @all modifier on the depot path:

    gp clone //depot/project@all

The environment and credentials are saved in the global Git config, please
use `--reset-credentials` to clear the config values:

    gp clone //depot/project --reset-credentials

### gp rebase

A common working pattern is to fetch the latest changes from the p4 depot and merge them with local uncommitted changes. Often, the p4 repository is the ultimate location for all code, thus a rebase workflow makes sense. `gp rebase` does `git p4 sync` followed by `git rebase` to move local commits on top of updated p4 changes.

### gp submit

To submit all changes that are in the current Git branch but not in the p4/master branch, use:

    $ git p4 submit

### gp shelve

To shelve all changes that are in the current Git branch, use:

    $ git p4 shelve

## Contribution

Please use Gradle build script for the testing the package before submitting the code:

    $ gradle tests

To install the application locally please use the `install` task:

    $ gradle install
