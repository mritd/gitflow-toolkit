# GitFlow ToolKit

> GitFlow Toolkit is a gitflow commit tool written by go, used to standardize the format of git commit message and quickly create gitflow branches,
> It should be noted that GitFlow Toolkit currently only supports the generation of the commit message style of the [Angular community specification](https://docs.google.com/document/d/1QrDFcIiPjSLDn3EL15IJygNPiHORgU1_OOAqWjiDU5Y/edit#heading=h.greljkmo14y0).

<p align="center">
  <img src="https://raw.githubusercontent.com/mritd/gitflow-toolkit/master/gitflow-toolkit.svg?sanitize=true" width="1200" alt="gitflow-toolkit demo">
</p>

## Installation

Just download the latest version from the Release page and execute the install command:

```sh
export VERSION='v2.0.0'

# download bin file
wget https://github.com/mritd/gitflow-toolkit/releases/download/${VERSION}/gitflow-toolkit_darwin_amd64

# add permissions
chmod +x gitflow-toolkit_darwin_amd64

# install
sudo ./gitflow-toolkit_darwin_amd64 install
```

After the installation is complete, you can delete the bin file.

If the go language development environment is installed locally, you can install it through the `go get` command:

```sh
go get -u github.com/mritd/gitflow-toolkit/v2
```

## Comands

| cmd | desc |
| --- | --- |
| `git ci` | Enter commit message interactively |
| `git ps` | Push the current branch to the remote |
| `git feat NAME` | Checkout a new branch from the current branch (`feat/NAME`) |
| `git fix NAME` | `git checkout -b fix/NAME` |
| `git hotfix NAME` | `git checkout -b hotfix/NAME` |
| `git docs NAME` | `git checkout -b docs/NAME` |
| `git style NAME` | `git checkout -b style/NAME` |
| `git refactor NAME` | `git checkout -b refactor/NAME` |
| `git chore NAME` | `git checkout -b chore/NAME` |
| `git perf NAME` | `git checkout -b perf/NAME` |
| `git style NAME` | `git checkout -b style/NAME` |

