package build

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	git "github.com/libgit2/git2go/v33"
	bocker "github.com/seppo0010/bocker/protocol"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
)

func getConfig(t *testing.T) (*bocker.BuildRequest, *shared.Config, func(), error) {
	cwdPath, err := os.MkdirTemp("", "bocker-test")
	assert.Nil(t, err)

	bockerTmpdir, err := os.MkdirTemp("", "bocker-test")
	assert.Nil(t, err)

	bockerBuildPath, err := os.MkdirTemp("", "bocker-build-path")
	assert.Nil(t, err)

	return &bocker.BuildRequest{
			CwdPath:  cwdPath,
			FilePath: path.Join(cwdPath, "Bockerfile"),
		}, &shared.Config{
			BockerPath: bockerTmpdir,
			BuildPath:  bockerBuildPath,
		}, func() {
			os.RemoveAll(cwdPath)
			os.RemoveAll(bockerTmpdir)
			os.RemoveAll(bockerBuildPath)
		}, nil
}

func getMetadata(t *testing.T, repo *git.Repository, tree *git.Tree) map[string]interface{} {
	entry, err := tree.EntryByPath("metadata")
	assert.Nil(t, err)
	obj, err := repo.Lookup(entry.Id)
	assert.Nil(t, err)
	blob, err := obj.AsBlob()
	assert.Nil(t, err)
	var data map[string]interface{}
	json.Unmarshal(blob.Contents(), &data)
	return data
}

func TestFileDoesNotExist(t *testing.T) {
	tmpdir, err := os.MkdirTemp("", "bocker-test")
	assert.Nil(t, err)
	defer os.RemoveAll(tmpdir)

	in, conf, cleanup, err := getConfig(t)
	assert.Nil(t, err)
	defer cleanup()

	_, err = Run(in, conf)
	assert.Equal(t, err, ErrFailedToOpen)
}

func TestInvalidFileContent(t *testing.T) {
	in, conf, cleanup, err := getConfig(t)
	assert.Nil(t, err)
	defer cleanup()

	f, err := os.Create(in.FilePath)
	assert.Nil(t, err)
	f.Close()

	_, err = Run(in, conf)
	assert.Equal(t, err, ErrFailedToParse)
}

func TestInvalidInstruction(t *testing.T) {
	in, conf, cleanup, err := getConfig(t)
	assert.Nil(t, err)
	defer cleanup()

	f, err := os.Create(in.FilePath)
	assert.Nil(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("FOO\n")
	assert.Nil(t, err)
	err = f.Close()
	assert.Nil(t, err)

	_, err = Run(in, conf)
	assert.Equal(t, err, ErrFailedToParseInstruction)
}

func TestUnsupportedInstruction(t *testing.T) {
	in, conf, cleanup, err := getConfig(t)
	assert.Nil(t, err)
	defer cleanup()

	f, err := os.Create(in.FilePath)
	assert.Nil(t, err)
	defer os.Remove(f.Name())
	_, err = f.WriteString("EXPOSE 123\n")
	assert.Nil(t, err)
	err = f.Close()
	assert.Nil(t, err)

	_, err = Run(in, conf)
	assert.Equal(t, err, ErrUnsupportedCommand)
}

func copyFile(t *testing.T, in *bocker.BuildRequest, conf *shared.Config, src, target, content string) string {
	f, err := os.Create(path.Join(in.CwdPath, src))
	assert.Nil(t, err)
	f.WriteString(content)
	f.Close()

	f, err = os.Create(in.FilePath)
	assert.Nil(t, err)
	_, err = f.WriteString(fmt.Sprintf("COPY %s %s\n", src, target))
	assert.Nil(t, err)
	err = f.Close()
	assert.Nil(t, err)
	return f.Name()
}

func TestCopy(t *testing.T) {
	in, conf, cleanup, err := getConfig(t)
	assert.Nil(t, err)
	defer cleanup()
	name := copyFile(t, in, conf, "foo", "bar", "baz")
	_, err = Run(in, conf)
	defer os.Remove(name)
	assert.Nil(t, err)

	f, err := os.Open(path.Join(conf.BuildPath, "bar"))
	assert.Nil(t, err)
	content, err := ioutil.ReadAll(f)
	assert.Nil(t, err)
	assert.Equal(t, content, []byte("baz"))
}

func TestCreateTree(t *testing.T) {
	hook := test.NewGlobal()
	log.SetLevel(log.DebugLevel)

	in, conf, cleanup, err := getConfig(t)
	assert.Nil(t, err)
	defer cleanup()
	name := copyFile(t, in, conf, "foo", "bar", "baz")
	treeId, err := Run(in, conf)
	defer os.Remove(name)
	assert.Nil(t, err)

	fileId := ""
	for _, entry := range hook.Entries {
		if entry.Message == "inserting file" {
			fileId = entry.Data["gitId"].(string)
		}
	}

	repo, err := git.OpenRepository(path.Join(conf.BockerPath, "repository"))
	assert.Nil(t, err)
	fileOid, err := git.NewOid(fileId)
	assert.Nil(t, err)
	obj, err := repo.Lookup(fileOid)
	assert.Nil(t, err)
	blob, err := obj.AsBlob()
	assert.Nil(t, err)
	assert.Equal(t, string(blob.Contents()), "baz")

	treeOid, err := git.NewOid(treeId)
	assert.Nil(t, err)
	obj, err = repo.Lookup(treeOid)
	assert.Nil(t, err)
	tree, err := obj.AsTree()
	assert.Nil(t, err)
	assert.Equal(t, tree.EntryCount(), uint64(1))

	treeHash := getMetadata(t, repo, tree)["TreeHash"].(string)

	oid, err := git.NewOid(treeHash)
	assert.Nil(t, err)
	dataTreeObject, err := repo.Lookup(oid)
	assert.Nil(t, err)
	dataTree, err := dataTreeObject.AsTree()
	assert.Nil(t, err)
	entry, err := dataTree.EntryByPath("bar")
	assert.Nil(t, err)
	assert.Equal(t, entry.Id, fileOid)
}
