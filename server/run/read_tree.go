package run

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	git "github.com/libgit2/git2go/v33"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
)

var ErrInvalidTreeHash = errors.New("invalid tree hash")
var ErrFailedToCopyTreeFiles = errors.New("failed to copy tree files")
var ErrTagNotFound = errors.New("tag not found")
var ErrFailedToGetTag = errors.New("failed to get tag")

func getTree(repo *git.Repository, hash string) (*git.Tree, error) {
	oid, err := git.NewOid(hash)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"hash":  hash,
		}).Error(ErrInvalidTreeHash.Error())
		return nil, ErrInvalidTreeHash
	}
	dataTreeObject, err := repo.Lookup(oid)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"hash":  hash,
		}).Error(ErrInvalidTreeHash.Error())
		return nil, ErrInvalidTreeHash
	}
	tree, err := dataTreeObject.AsTree()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"hash":  hash,
		}).Error(ErrInvalidTreeHash.Error())
		return nil, ErrInvalidTreeHash
	}
	return tree, nil
}

func getTag(repo *git.Repository, tag string) (string, error) {
	refPath := fmt.Sprintf("refs/tags/%s", tag)
	ref, err := repo.References.Lookup(refPath)
	if err != nil {
		if err.Error() == fmt.Sprintf("reference '%s' not found", refPath) {
			return "", ErrTagNotFound
		}
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrFailedToGetTag.Error())
		return "", ErrFailedToGetTag
	}
	obj, err := ref.Peel(git.ObjectTree)
	if err != nil {
		// obj is not a tree? can still be a hash, right?
		return "", nil
	}
	return obj.Id().String(), nil
}

func readMetadata(repo *git.Repository, hashOrTag string) (*shared.Metadata, error) {
	hash, err := getTag(repo, hashOrTag)
	if err != nil && err != ErrTagNotFound {
		return nil, err
	}
	var h string
	if hash != "" {
		h = hash
	} else {
		h = hashOrTag
	}
	tree, err := getTree(repo, h)
	if err != nil {
		return nil, err
	}

	entry, err := tree.EntryByPath("metadata")
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"hash":  hash,
		}).Error(ErrInvalidTreeHash.Error())
		return nil, ErrInvalidTreeHash
	}
	obj, err := repo.Lookup(entry.Id)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"hash":  hash,
		}).Error(ErrInvalidTreeHash.Error())
		return nil, ErrInvalidTreeHash
	}
	blob, err := obj.AsBlob()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"hash":  hash,
		}).Error(ErrInvalidTreeHash.Error())
		return nil, ErrInvalidTreeHash
	}
	var metadata *shared.Metadata
	err = json.Unmarshal(blob.Contents(), &metadata)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
			"hash":  hash,
		}).Error(ErrInvalidTreeHash.Error())
		return nil, ErrInvalidTreeHash
	}
	return metadata, nil
}

func copyTreeFiles(repo *git.Repository, treeHash, runPath string) error {
	tree, err := getTree(repo, treeHash)
	if err != nil {
		return err
	}
	var walkErr error
	err = tree.Walk(func(_ string, entry *git.TreeEntry) error {
		filePath := entry.Name
		if filePath == "" {
			return nil
		}
		walkErr = os.MkdirAll(path.Join(
			runPath,
			path.Dir(filePath),
		), os.ModePerm)
		if walkErr != nil {
			return walkErr
		}

		var obj *git.Object
		obj, walkErr = repo.Lookup(entry.Id)
		if walkErr != nil {
			return walkErr
		}
		var blob *git.Blob
		blob, walkErr = obj.AsBlob()
		if walkErr != nil {
			return walkErr
		}

		var f *os.File
		f, walkErr = os.Create(path.Join(runPath, filePath))
		if walkErr != nil {
			return walkErr
		}
		_, walkErr = f.Write(blob.Contents())
		if walkErr != nil {
			return walkErr
		}
		if entry.Filemode == git.FilemodeBlobExecutable {
			walkErr = f.Chmod(0744)
		}
		walkErr = f.Close()
		if walkErr != nil {
			return walkErr
		}
		return nil
	})
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrFailedToCopyTreeFiles.Error())
	}
	if walkErr != nil {
		log.WithFields(log.Fields{
			"error": walkErr.Error(),
		}).Error(ErrFailedToCopyTreeFiles.Error())
	}
	return nil
}
