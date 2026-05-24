package span

import "fmt"

var AppName = "span"
var AppVersion = "unknown"

var UserAgent = fmt.Sprintf("%s/%s", AppName, AppVersion)
