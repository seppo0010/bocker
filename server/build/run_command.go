package build

import (
	"errors"
	"os"
	"os/exec"
	"strings"

	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
)

var ErrCommandFailed = errors.New("command failed")

func runCommand(conf *shared.Config, args []string) error {
	// FIXME: this is probably wrong
	cmd := exec.Command("bash", "-c", strings.Join(args, " "))
	cmd.Dir = conf.BuildPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.WithFields(log.Fields{
			"command": cmd,
			"error":   err.Error(),
		}).Error(ErrCommandFailed.Error())
		return ErrCommandFailed
	}
	return nil
}
