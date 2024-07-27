package job

// 检查用户是否有权删除该评论，如下情况之一可以删
// 1. 用户是该评论的作者
// 2. 用户是该评论对象的作者

// if uid != reply.Uid {
// 	resp, err := external.GetNoter().IsUserOwnNote(ctx, &notesdk.IsUserOwnNoteReq{
// 		Uid:     uid,
// 		NoteIds: []uint64{reply.Oid},
// 	})
// 	if err != nil {
// 		logx.Errorf("check IsUserOwnNote err: %v, rid: %d, uid: %d", err, rid, uid)
// 		return global.ErrInternal
// 	}

// 	if len(resp.GetResult()) < 1 {
// 		logx.Errorf("check IsUserOwnNote result len is 0: rid: %d, uid: %d", rid, uid)
// 		return global.ErrInternal
// 	}

// 	if !resp.GetResult()[0] {
// 		return global.ErrPermDenied
// 	}
// }

// // 是否是主评论 如果为主评论 需要一并删除所有子评论
// if isRootReply(reply.RootId, reply.ParentId) {

// } else {
// 	// 只需要删除评论本身

// }
