package span

import "fmt"

var Name string = "span"
var Version string = "unkown"

func ShowVersion() {
	fmt.Printf("%s/%s\n", Name, Version)
}
