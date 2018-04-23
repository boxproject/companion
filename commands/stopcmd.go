package commands

import (
	"os/exec"

	"gopkg.in/urfave/cli.v1"
)

func StopCmd(_ *cli.Context) error {
	_, err := exec.Command("sh", "-c", "pkill -SIGINT companion").Output()
	return err
}
