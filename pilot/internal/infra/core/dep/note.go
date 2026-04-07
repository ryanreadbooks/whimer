package dep

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/config"

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
	conn := xgrpc.NewRecoverableClientConn(c.Backend.Note)
	noteCreator = notev1.NewNoteCreatorServiceClient(conn)
	noteFeed = notev1.NewNoteFeedServiceClient(conn)
	noteInteract = notev1.NewNoteInteractServiceClient(conn)
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
