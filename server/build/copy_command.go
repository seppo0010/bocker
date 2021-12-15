package build

import (
	"os"
	"path"
	"strings"

	pb "github.com/seppo0010/bocker/protocol"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
)

func copyCommand(
	in *pb.BuildRequest,
	conf *shared.Config,
	from []string,
	to string,
) error {
	log.WithFields(log.Fields{
		"from": from,
		"to":   to,
	}).Info("copying files")
	for _, filepath := range from {
		var target string
		source := path.Join(in.GetCwdPath(), filepath)
		if strings.HasSuffix(to, "/") {
			target = path.Join(conf.BuildPath, to, path.Base(filepath))
		} else {
			target = path.Join(conf.BuildPath, to)
		}
		fi, err := os.Stat(source)
		if err != nil {
			log.WithFields(log.Fields{
				"error":  err.Error(),
				"source": source,
			}).Error(ErrFailedToCopyFile.Error())
			return ErrFailedToCopyFile
		}
		switch mode := fi.Mode(); {
		case mode.IsDir():
			err = CopyDir(source, target)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err.Error(),
					"source": source,
					"target": target,
				}).Error(ErrFailedToCopyFile.Error())
				return ErrFailedToCopyFile
			}
		case mode.IsRegular():
			err = CopyFile(source, target)
			if err != nil {
				log.WithFields(log.Fields{
					"error":  err.Error(),
					"source": source,
					"target": target,
				}).Error(ErrFailedToCopyFile.Error())
				return ErrFailedToCopyFile
			}
		}
	}
	return nil
}
