package build

import (
	"fmt"
	"os"
	"path"
	"testing"

	git "github.com/libgit2/git2go/v33"
	"github.com/stretchr/testify/assert"
)

func TestMaintainer(t *testing.T) {
	in, conf, cleanup, err := getConfig(t)
	assert.Nil(t, err)
	defer cleanup()

	expectedMaintainer := "example@example.com"
	f, err := os.Create(in.FilePath)
	assert.Nil(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString(fmt.Sprintf("MAINTAINER %s\n", expectedMaintainer))
	assert.Nil(t, err)
	err = f.Close()
	assert.Nil(t, err)

	treeId, err := Run(in, conf)
	assert.Nil(t, err)

	repo, err := git.OpenRepository(path.Join(conf.BockerPath, "repository"))
	assert.Nil(t, err)
	treeOid, err := git.NewOid(treeId)
	assert.Nil(t, err)
	obj, err := repo.Lookup(treeOid)
	assert.Nil(t, err)
	tree, err := obj.AsTree()
	assert.Nil(t, err)

	maintainer := getMetadata(t, repo, tree)["Maintainer"].(string)
	assert.Equal(t, maintainer, expectedMaintainer)
}
