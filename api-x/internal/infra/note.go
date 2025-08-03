package infra

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"

	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

var (
	// 笔记服务
	noteCreator  notev1.NoteCreatorServiceClient
	noteFeed     notev1.NoteFeedServiceClient
	noteInteract notev1.NoteInteractServiceClient

	noteIdObfuscate obfuscate.Obfuscate
	tagIdObfuscate  obfuscate.Obfuscate
)

func InitNote(c *config.Config) {
	initNoteBackend(c)
	initNoteObfuscate(c)
	initTagObfuscate(c)
}

func initNoteBackend(c *config.Config) {
	noteCreator = xgrpc.NewRecoverableClient(c.Backend.Note,
		notev1.NewNoteCreatorServiceClient,
		func(cc notev1.NoteCreatorServiceClient) { noteCreator = cc })
	noteFeed = xgrpc.NewRecoverableClient(c.Backend.Note,
		notev1.NewNoteFeedServiceClient,
		func(cc notev1.NoteFeedServiceClient) { noteFeed = cc })
	noteInteract = xgrpc.NewRecoverableClient(c.Backend.Note,
		notev1.NewNoteInteractServiceClient,
		func(cc notev1.NoteInteractServiceClient) { noteInteract = cc })
}

func NoteCreatorServer() notev1.NoteCreatorServiceClient {
	return noteCreator
}

func NoteInteractServer() notev1.NoteInteractServiceClient {
	return noteInteract
}

func NoteFeedServer() notev1.NoteFeedServiceClient {
	return noteFeed
}

func GetNoteIdObfuscate() obfuscate.Obfuscate {
	return noteIdObfuscate
}

func GetTagIdObfuscate() obfuscate.Obfuscate {
	return tagIdObfuscate
}

// init note id obfuscator
func initNoteObfuscate(c *config.Config) {
	noteIdObfuscate, err = obfuscate.NewConfuser(c.Obfuscate.Note.Options()...)
	if err != nil {
		panic(fmt.Errorf("init note obfuscate: %w", err))
	}
}

func initTagObfuscate(c *config.Config) {
	tagIdObfuscate, err = obfuscate.NewConfuser(c.Obfuscate.Tag.Options()...)
	if err != nil {
		panic(fmt.Errorf("init tag obfuscate: %w", err))
	}
}
