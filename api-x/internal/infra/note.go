package infra

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"

	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"
)

var (
	// 笔记服务
	noteCreator  notev1.NoteCreatorServiceClient
	noteFeed     notev1.NoteFeedServiceClient
	noteInteract notev1.NoteInteractServiceClient
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

// init note id obfuscator
func initNoteObfuscate(c *config.Config) {
	noteid.InitNoteIdObfuscate(c.Obfuscate.Note.Options()...)
}

func initTagObfuscate(c *config.Config) {
	noteid.InitTagIdObfuscate(c.Obfuscate.Tag.Options()...)
}
