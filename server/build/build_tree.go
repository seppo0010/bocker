package build

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	git "github.com/libgit2/git2go/v33"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
)

func getBuildTree(conf *shared.Config, repo *git.Repository) (*git.Tree, error) {
	tb, err := repo.TreeBuilder()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrFailedToGetTreeID.Error())
		return nil, ErrFailedToGetTreeID
	}

	err = filepath.WalkDir(conf.BuildPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// FIXME
			panic(err.Error())
		}
		if d.IsDir() {
			return nil
		}

		if !strings.HasPrefix(path, conf.BuildPath) {
			panic("Expected all files to be descendants of the build path")
		}

		f, err := os.Open(path)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"path":  path,
			}).Error("failed to open file")
			return ErrFailedToGetTreeID
		}
		defer f.Close()

		content, err := ioutil.ReadAll(f)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"path":  path,
			}).Error("failed to read file content")
			return ErrFailedToGetTreeID
		}

		fileOid, err := repo.CreateBlobFromBuffer(content)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
				"path":  path,
			}).Error("failed to add file to tree")
			return ErrFailedToGetTreeID
		}

		stat, err := f.Stat()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("failed to stat file")
			return ErrFailedToGetTreeID
		}

		gitMode := git.FilemodeBlob
		if stat.Mode()&0100 != 0 {
			gitMode = git.FilemodeBlobExecutable
		}

		target := path[len(conf.BuildPath)+1:]
		tb.Insert(
			target,
			fileOid,
			gitMode,
		)
		log.WithFields(log.Fields{
			"sourcePath": path,
			"target":     target,
			"gitId":      fileOid.String(),
		}).Debug("inserting file")
		return nil
	})

	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrFailedToGetTreeID.Error())
		return nil, ErrFailedToGetTreeID
	}

	treeOid, err := tb.Write()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrFailedToGetTreeID.Error())
		return nil, ErrFailedToGetTreeID
	}
	log.WithFields(log.Fields{
		"gitId": treeOid.String(),
	}).Debug("writing tree")

	treeObject, err := repo.Lookup(treeOid)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrFailedToGetTreeID.Error())
		return nil, ErrFailedToGetTreeID
	}
	tree, err := treeObject.AsTree()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrFailedToGetTreeID.Error())
		return nil, ErrFailedToGetTreeID
	}
	return tree, nil
}
