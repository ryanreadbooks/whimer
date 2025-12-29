package procedure

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/data/event"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

var (
	_ Procedure     = (*PublishProcedure)(nil)
	_ AutoCompleter = (*PublishProcedure)(nil)
)

// 笔记发布
//
// 流水线中的最后一步发布流程
type PublishProcedure struct {
	noteBiz        *biz.NoteBiz
	noteCreatorBiz *biz.NoteCreatorBiz
	noteEventBus   *event.NoteEventBus
}

func NewPublishProcedure(bizz *biz.Biz) *PublishProcedure {
	return &PublishProcedure{
		noteBiz:        bizz.Note,
		noteCreatorBiz: bizz.Creator,
		noteEventBus:   bizz.Data().NoteEventBus,
	}
}

func (p *PublishProcedure) Type() model.ProcedureType {
	return model.ProcedureTypePublish
}

func (p *PublishProcedure) BeforeExecute(ctx context.Context, note *model.Note) (bool, error) {
	err := p.noteCreatorBiz.TransferNoteStateToPublished(ctx, note)
	if err != nil {
		return false, xerror.Wrapf(err, "publish procedure set note state published failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return true, nil
}

func (p *PublishProcedure) doExecute(ctx context.Context, note *model.Note) (string, error) {
	if note.Privacy == model.PrivacyPrivate {
		// TODO 这里需要分辨情况
		// 这里要按照privacy来处理吗 还是留给事件处理方来处理
		return "", nil
	}

	// 获取完整的note数据 包括asset资源数据
	// 此处的note是流程发起早期的数据 有些异步生成的数据包含在内
	fullNote, err := p.noteBiz.GetNote(ctx, note.NoteId)
	if err != nil {
		return "", xerror.Wrapf(err, "publish procedure get note failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	// 必须补全ext数据
	err = p.noteBiz.AssembleNotesExt(ctx, fullNote.AsSlice())
	if err != nil {
		return "", xerror.Wrapf(err, "publish procedure assemble notes ext failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	err = p.noteEventBus.NotePublished(ctx, fullNote)
	if err != nil {
		return "", xerror.Wrapf(err, "publish procedure note published event failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return uuid.NewUUID().String(), nil
}

// 广播笔记发布事件
func (p *PublishProcedure) Execute(ctx context.Context, note *model.Note) (string, error) {
	if note.Privacy == model.PrivacyPrivate {
		return "", nil
	}

	return p.doExecute(ctx, note)
}

// 消息队列信息发送成功
func (p *PublishProcedure) OnSuccess(ctx context.Context, result *ProcedureResult) (bool, error) {
	// 简单记录
	xlog.Msgf("publish procedure on success completed, note(%d) is published to events", result.NoteId).
		Extras("note_id", result.NoteId, "task_id", result.TaskId).
		Infox(ctx)

	return true, nil
}

func (p *PublishProcedure) OnFailure(ctx context.Context, result *ProcedureResult) (bool, error) {
	// 简单记录
	xlog.Msgf("publish procedure on failure completed, note(%d) is not published to events", result.NoteId).
		Extras("note_id", result.NoteId, "task_id", result.TaskId).
		Infox(ctx)

	return true, nil
}

func (p *PublishProcedure) PollResult(ctx context.Context, record *biz.ProcedureRecord) (PollState, any, error) {
	// TODO 轮询发布结果
	return PollStateSuccess, nil, nil
}

func (p *PublishProcedure) Retry(ctx context.Context, record *biz.ProcedureRecord) (string, error) {
	note, err := p.noteBiz.GetNoteCoreWithoutCache(ctx, record.NoteId)
	if err != nil {
		xlog.Msg("asset procedure retry get note failed").
			Err(err).
			Extras("record_id", record.Id, "note_id", record.NoteId).
			Errorx(ctx)
		return "", err
	}

	return p.executeForRetry(ctx, note)
}

func (p *PublishProcedure) executeForRetry(ctx context.Context,
	note *model.Note,
) (string, error) {
	// 重新尝试写入消息队列
	taskId, err := p.doExecute(ctx, note)
	if err != nil {
		return "", xerror.Wrapf(err, "publish procedure execute for retry failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return taskId, nil
}

func (p *PublishProcedure) ObAbort(ctx context.Context, note *model.Note, taskId string) error {
	return nil
}

// 自动成功 只要Execute成功了就是视为成功了
func (p *PublishProcedure) AutoComplete(ctx context.Context,
	note *model.Note, taskId string,
) (success, autoComplete bool, arg any) {
	return true, true, nil
}
