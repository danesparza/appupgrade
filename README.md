# appupgrade [![CircleCI](https://circleci.com/gh/danesparza/appupgrade.svg?style=shield)](https://circleci.com/gh/danesparza/appupgrade)
A sidecar REST service to upgrade a local app to the latest version.  Uses debian packages (*.deb files) and github as its system of record for version information

Deprecated:  [Just host a PPA repository, instead](https://assafmo.github.io/2019/05/02/ppa-repo-hosted-on-github.html). 

Its port can be configured, but by default the service documentation resides at `http://<url>:3007/v1/swagger/`
