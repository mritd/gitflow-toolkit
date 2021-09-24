# GitFlow ToolKit

> GitFlow Toolkit is a gitflow commit tool written by go, used to standardize the format of git commit message and quickly create gitflow branches,
> It should be noted that GitFlow Toolkit currently only supports the generation of the commit message style of the [Angular community specification](https://docs.google.com/document/d/1QrDFcIiPjSLDn3EL15IJygNPiHORgU1_OOAqWjiDU5Y/edit#heading=h.greljkmo14y0).

### The overall UI has been rewritten since v2.1.0. The current v2.1.0 version is still under development, but it is basically available.


| | | 
|:-------------------------:|:-------------------------:|
|<img width="2500" src="https://user-images.githubusercontent.com/13043245/134647305-a1df0023-972b-48c3-a6bf-668e96094df9.gif"> Install|<img width="2500" src="https://user-images.githubusercontent.com/13043245/134646600-976f01b4-6000-41e7-8389-0d0e761e15c9.gif"> Uninstall|
|<img width="2500" src="https://user-images.githubusercontent.com/13043245/134485491-993ef0cb-7438-4c42-9a2e-16db05503a0b.gif"> Commit Success|<img width="2500" src="https://user-images.githubusercontent.com/13043245/134485537-6375d280-10d2-4475-a834-7d0ad72248aa.gif"> Commit Failed|
|<img width="2500" src="https://user-images.githubusercontent.com/13043245/134485533-3a01d3be-0912-45cb-9e63-d343a7bad847.gif"> Push Success|<img width="2500" src="https://user-images.githubusercontent.com/13043245/134485503-f7de0493-6d2d-403d-aa4d-79a62a83c048.gif"> Push Failed|
|<img width="2500" src="https://user-images.githubusercontent.com/13043245/134485549-5ee7853d-1cc7-4a0f-b083-03514045f8eb.gif"> Create Branch||

## Installation

Just download the latest version from the Release page and execute the `install` command:

```sh
export VERSION='v2.1.0'

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
go install github.com/mritd/gitflow-toolkit/v2@latest
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

