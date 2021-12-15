package run

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"path"

	pb "github.com/seppo0010/bocker/protocol"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
)

var ErrCannotCreateRunDirectory = errors.New("cannot create run directory")
var ErrRunFailed = errors.New("run failed")

func bufferNotifier(notify func(b []byte) error) (io.ReadCloser, io.Writer) {
	read, write := io.Pipe()
	go func() {
		for {
			b := make([]byte, 1024)
			size, err := io.ReadAtLeast(read, b, 1)
			if err != nil {
				break
			}
			err = notify(b[:size])
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Error("failed to notify")
				break
			}
		}
	}()
	return read, write
}

func Run(in *pb.RunRequest, conf *shared.Config, outchan chan<- *pb.ExecReply) error {
	defer close(outchan)
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
	var readStdout, readStderr io.ReadCloser
	readStdout, cmd.Stdout = bufferNotifier(func(b []byte) error {
		outchan <- &pb.ExecReply{
			Stdout:   b,
			ExitCode: ^uint32(0),
		}
		return nil
	})
	readStderr, cmd.Stderr = bufferNotifier(func(b []byte) error {
		outchan <- &pb.ExecReply{
			Stderr:   b,
			ExitCode: ^uint32(0),
		}
		return nil
	})
	err = cmd.Run()
	readStderr.Close()
	readStdout.Close()
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
