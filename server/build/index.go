package build

import (
	"errors"
	"os"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	pb "github.com/seppo0010/bocker/protocol"
	"github.com/seppo0010/bocker/shared"
	log "github.com/sirupsen/logrus"
)

var ErrFailedToOpen = errors.New("failed to open Bockerfile")
var ErrFailedToParse = errors.New("failed to parse Bockerfile")
var ErrFailedToParseInstruction = errors.New("failed to parse Bockerfile instruction")
var ErrUnsupportedCommand = errors.New("command not supported")
var ErrFailedToGetRepository = errors.New("failed to get repository")
var ErrFailedToGetTreeID = errors.New("failed to get empty tree id")
var ErrFailedToCopyFile = errors.New("failed to copy file")

func Run(in *pb.BuildRequest, conf *shared.Config) (string, error) {
	tag := in.Tag
	filename := in.FilePath
	f, err := os.Open(filename)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error(ErrFailedToOpen.Error())
		return "", ErrFailedToOpen
	}
	res, err := parser.Parse(f)
	if err != nil {
		log.WithFields(log.Fields{
			"filename": filename,
			"error":    err.Error(),
		}).Error(ErrFailedToParse.Error())
		return "", ErrFailedToParse
	}
	log.WithFields(log.Fields{
		"parseResult": res,
	}).Debug("parsed bockerfile")
	// FIXME: avoid double parse
	for _, node := range res.AST.Children {
		instruction, err := instructions.ParseInstruction(node)
		if err != nil {
			log.WithFields(log.Fields{
				"filename": filename,
				"error":    err.Error(),
				"node":     node.Original,
			}).Error(ErrFailedToParseInstruction)
			return "", ErrFailedToParseInstruction
		}
		switch instruction.(type) {
		case *instructions.CopyCommand:
		case *instructions.CmdCommand:
		case *instructions.RunCommand:
		case *instructions.MaintainerCommand:
			break
		default:
			log.WithFields(log.Fields{
				"filename": filename,
				"node":     node.Original,
			}).Error(ErrUnsupportedCommand.Error())
			return "", ErrUnsupportedCommand
		}
	}

	repo, err := shared.GetRepository(conf)
	if err != nil {
		return "", err
	}

	tree, err := getBuildTree(conf, repo)
	if err != nil {
		return "", err
	}
	metadata := &shared.Metadata{
		TreeHash: tree.Object.Id().String(),
	}
	head, err := getMetadataTree(conf, repo, metadata)
	if err != nil {
		return "", err
	}
	for _, node := range res.AST.Children {
		instruction, _ := instructions.ParseInstruction(node)
		maybeDirty := false
		switch v := instruction.(type) {
		case *instructions.CopyCommand:
			err = copyCommand(in, conf, v.SourcesAndDest.SourcePaths, v.SourcesAndDest.DestPath)
			if err != nil {
				return "", err
			}
			maybeDirty = true
		case *instructions.MaintainerCommand:
			metadata.Maintainer = v.Maintainer
		case *instructions.CmdCommand:
			metadata.Command = v.CmdLine
		case *instructions.RunCommand:
			err = runCommand(conf, v.CmdLine)
			if err != nil {
				return "", err
			}
			maybeDirty = true
		default:
			panic("unreachable")
		}
		if maybeDirty {
			tree, err = getBuildTree(conf, repo)
			if err != nil {
				return "", err
			}
			metadata.TreeHash = tree.Object.Id().String()
		}
		head, err = getMetadataTree(conf, repo, metadata)
		if err != nil {
			return "", err
		}
	}
	hash := head.Object.Id().String()
	if tag != "" {
		_, err := repo.Tags.CreateLightweight(tag, head, true)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Warn("failed to create tag")
		}
	}
	return hash, nil
}
