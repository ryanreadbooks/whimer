package srv

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/panjf2000/ants/v2"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
	"github.com/ryanreadbooks/whimer/search/internal/infra"
	noteindex "github.com/ryanreadbooks/whimer/search/internal/infra/esdao/index/note"
	"github.com/ryanreadbooks/whimer/search/internal/infra/kafkadao"
	"github.com/ryanreadbooks/whimer/search/pkg"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/encoding/protojson"
)

type DocumentService struct {
	pool *ants.Pool
}

func NewDocumentService() *DocumentService {
	pool, _ := ants.NewPool(100)
	d := &DocumentService{
		pool: pool,
	}

	return d
}

func (s *DocumentService) Close() {
	s.pool.Release()
}

func makeIndexNote(n *searchv1.Note) *noteindex.Note {
	indexTags := make([]*noteindex.NoteTag, 0, len(n.GetTagList()))
	for _, t := range n.GetTagList() {
		indexTags = append(indexTags, &noteindex.NoteTag{
			Id:    t.Id,
			Name:  t.Name,
			Ctime: t.Ctime,
		})
	}

	return &noteindex.Note{
		NoteId:   n.NoteId,
		Title:    n.Title,
		Desc:     n.Desc,
		CreateAt: n.CreateAt,
		UpdateAt: n.UpdateAt,
		Author: noteindex.NoteAuthor{
			Uid:      n.Author.Uid,
			Nickname: n.Author.Nickname,
		},
		TagList:    indexTags,
		AssetType:  pkg.NoteAssetConverter[n.GetAssetType()],
		Visibility: pkg.NoteVisibilityConverter[n.GetVisibility()],
	}
}

// 异步提交写入
func (s *DocumentService) AddNoteTagDocs(ctx context.Context, tags []*searchv1.NoteTag) error {
	indexTags := make([]*noteindex.NoteTag, 0, len(tags))
	for _, t := range tags {
		indexTags = append(indexTags, &noteindex.NoteTag{
			Id:    t.Id,
			Name:  t.Name,
			Ctime: t.Ctime,
		})
	}

	concurrent.ControllableExec(ctx, s.pool, indexTags, func(ctx context.Context, datas []*noteindex.NoteTag) error {
		err := infra.EsDao().NoteTagIndexer.BulkAdd(ctx, datas)
		if err != nil {
			xlog.Msg("document note tag add failed").Err(err).Errorx(ctx)
			return err
		}

		return nil
	})

	return nil
}

func (s *DocumentService) AddNoteDocs(ctx context.Context, notes []*searchv1.Note) error {
	// 异步写入kafka等待消费
	return infra.KafkaDao().NoteEventProducer.PutNoteAddEvent(ctx, notes)
}

func (s *DocumentService) DeleteNoteDocs(ctx context.Context, ids []string) error {
	return infra.KafkaDao().NoteEventProducer.PutNoteDeleteEvent(ctx, ids)
}

func (s *DocumentService) UpdateNoteDocLikeCount(ctx context.Context, reqs map[string]int64) error {
	return infra.KafkaDao().NoteEventProducer.PutNoteLikeEvent(ctx, reqs)
}

func (s *DocumentService) UpdateNoteDocCommentCount(ctx context.Context, reqs map[string]int64) error {
	return infra.KafkaDao().NoteEventProducer.PutNoteCommentEvent(ctx, reqs)
}

// 批量处理kafka消息
//
// 此处主要就是将批量的kafka通过es的bulk批量写入
func (s *DocumentService) DispatchNoteEvents(ctx context.Context, msgs []kafka.Message) error {
	var (
		errs     []error
		bulkReqs []noteindex.NoteAction
		tracer   = otel.Tracer("docsrv.dispatch")
	)

	bulkCtx, bulkSpan := tracer.Start(ctx, "docsrv.dispatch.note.event",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(semconv.MessagingBatchMessageCount(len(msgs))),
	)
	defer bulkSpan.End()

	xlog.Msgf("doc service dispatch note events handling %d msgs", len(msgs)).Infox(bulkCtx)

	for _, msg := range msgs {
		var ev kafkadao.NoteEvent
		err := json.Unmarshal(msg.Value, &ev)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		msgCtx := xkafka.ContextFromKafkaHeaders(msg.Headers)
		xlog.Msgf("handle message %v", ev.Type).Debugx(msgCtx)

		switch ev.Type {
		case kafkadao.NoteAddEvent:
			var req searchv1.Note
			err = protojson.Unmarshal(ev.Payload, &req)
			if err != nil {
				errs = append(errs, fmt.Errorf("add event protojson unmarshal: %w", err))
				continue
			}

			bulkReqs = append(bulkReqs, noteindex.NewNoteCreateAction(makeIndexNote(&req)))
		case kafkadao.NoteDeleteEvent:
			var noteId string
			err = json.Unmarshal(ev.Payload, &noteId)
			if err != nil {
				errs = append(errs, fmt.Errorf("delete event json unmarshal: %w", err))
				continue
			}

			bulkReqs = append(bulkReqs, noteindex.NewNoteDeleteAction(noteId))
		case kafkadao.NoteLikeEvent:
			var lv kafkadao.NoteLikeEventPayload
			err = json.Unmarshal(ev.Payload, &lv)
			if err != nil {
				errs = append(errs, fmt.Errorf("like event json unamrshal: %w", err))
				continue
			}
		case kafkadao.NoteCommentEvent:
			var cv kafkadao.NoteCommentEventPayload
			err = json.Unmarshal(ev.Payload, &cv)
			if err != nil {
				errs = append(errs, fmt.Errorf("comment event json unamrshal: %w", err))
				continue
			}

		default:
			xlog.Msg("unsupported note event types").Extras("msg.topic", msg.Topic, "msg.key", msg.Key).Errorx(msgCtx)
			continue
		}

		bulkSpan.AddLink(trace.LinkFromContext(msgCtx,
			attribute.KeyValue{
				Key:   attribute.Key("messaging.kafka.message.topic"),
				Value: attribute.StringValue(msg.Topic),
			},
			semconv.MessagingKafkaConsumerGroup(kafkadao.EsNoteTopicGroup),
			semconv.MessagingKafkaMessageKey(string(msg.Key)),
			semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
		))
	}

	if len(errs) != 0 {
		xlog.Msg("doc service dispatch note events has errors").Err(errors.Join(errs...)).Errorx(bulkCtx)
	}

	if len(bulkReqs) > 0 {
		err := infra.EsDao().NoteIndexer.BulkRequest(bulkCtx, bulkReqs)
		if err != nil {
			bulkSpan.SetStatus(codes.Error, err.Error())
			return xerror.Wrapf(err, "doc service note indexer bulk failed").WithCtx(bulkCtx)
		}
		bulkSpan.SetStatus(codes.Ok, "docsrv dispatch note events done")
	}

	return nil
}
