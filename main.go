// @title appupgrade
// @version 1.0
// @description A sidecar REST service to upgrade a local app to the latest version

// @license.name MIT
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath /v1
package main

import "github.com/danesparza/appupgrade/cmd"

func main() {
	cmd.Execute()
}
