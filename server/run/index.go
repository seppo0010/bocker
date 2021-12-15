package run

import (
	"errors"
	"os"
	"os/exec"
	"path"

	pb "github.com/seppo0010/bocker/protocol"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
)

var ErrCannotCreateRunDirectory = errors.New("cannot create run directory")
var ErrRunFailed = errors.New("run failed")

func Run(in *pb.RunRequest, conf *shared.Config) error {
	tag := in.Tag
	config := &shared.Config{
		BockerPath: path.Join(os.Getenv("HOME"), ".bocker"),
	}

	repo, err := shared.GetRepository(config)
	if err != nil {
		return err
	}
	metadata, err := readMetadata(repo, tag)
	if err != nil {
		return err
	}

	runPath, err := os.MkdirTemp("", "bocker-run")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrCannotCreateRunDirectory.Error())
		return ErrCannotCreateRunDirectory
	}
	defer os.RemoveAll(runPath)

	err = copyTreeFiles(repo, metadata.TreeHash, runPath)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrCannotCreateRunDirectory.Error())
		return ErrCannotCreateRunDirectory
	}

	cmd := exec.Command(metadata.Command[0], metadata.Command[1:]...)
	cmd.Dir = runPath
	// FIXME: send messages
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			// FIXME: send message
			log.WithFields(log.Fields{
				"code": err.ExitCode(),
			}).Info("exit code")
			return nil
		}
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrRunFailed.Error())
		return ErrRunFailed
	}
	return nil
}
