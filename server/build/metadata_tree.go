package build

import (
	"encoding/json"

	git "github.com/libgit2/git2go/v33"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
)

func getMetadataTree(conf *shared.Config, repo *git.Repository, metadata *shared.Metadata) (*git.Tree, error) {
	tb, err := repo.TreeBuilder()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error(ErrFailedToGetTreeID.Error())
		return nil, ErrFailedToGetTreeID
	}

	j, err := json.Marshal(metadata)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to create metadata json")
		return nil, ErrFailedToGetTreeID
	}
	fileOid, err := repo.CreateBlobFromBuffer(j)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("failed to add file to tree")
		return nil, ErrFailedToGetTreeID
	}
	tb.Insert(
		"metadata",
		fileOid,
		git.FilemodeBlob,
	)
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
