package shared

import (
	"errors"
	"path"

	git "github.com/libgit2/git2go/v33"
	log "github.com/sirupsen/logrus"
)

var ErrFailedToCreateRepository = errors.New("failed to create repository")

func GetRepository(conf *Config) (*git.Repository, error) {
	repoPath := path.Join(conf.BockerPath, "repository")
	repo, errOpen := git.OpenRepository(repoPath)
	if errOpen != nil {
		var err error
		repo, err = git.InitRepository(repoPath, true)
		if err != nil {
			log.WithFields(log.Fields{
				"errorOpen": errOpen.Error(),
				"errorInit": err.Error(),
			}).Error(ErrFailedToCreateRepository.Error())
			return nil, err
		}
	}
	return repo, nil
}
