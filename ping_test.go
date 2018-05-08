package pinggateway

import (
	"fmt"
	"os/exec"
	"testing"
)

func Test_gageway(t *testing.T) {

	if PingGateway(false) == false {
		fmt.Println("")
		exec.Command("pm-hibernate", "").Run()

	}
	fmt.Println("-")
}
