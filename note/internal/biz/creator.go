package biz

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"math"
	"time"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// 笔记相关
// 创作者相关
type NoteCreatorBiz struct {
	NoteBiz
	OssKeyGen *keygen.Generator
}

func NewNoteCreatorBiz() NoteCreatorBiz {
	b := NoteCreatorBiz{
		OssKeyGen: keygen.NewGenerator(
			keygen.WithBucket(config.Conf.Oss.Bucket),
			keygen.WithPrefix(config.Conf.Oss.Prefix),
			keygen.WithPrependBucket(true),
		),
	}

	return b
}

func (b *NoteCreatorBiz) CreateNote(ctx context.Context, note *model.CreateNoteRequest) (int64, error) {
	var (
		uid    = metadata.Uid(ctx)
		noteId int64
	)

	now := time.Now().Unix()
	newNote := &notedao.Note{
		Title:   note.Basic.Title,
		Desc:    note.Basic.Desc,
		Privacy: int8(note.Basic.Privacy),
		Owner:   uid,
	}

	var noteAssets = make([]*notedao.Asset, 0, len(note.Images))
	for _, img := range note.Images {
		imgMeta := model.NewAssetImageMeta(img.Width, img.Height, img.Format).String()
		noteAssets = append(noteAssets, &notedao.Asset{
			AssetKey:  img.FileId, // 包含桶名称
			AssetType: global.AssetTypeImage,
			NoteId:    noteId,
			CreateAt:  now,
			AssetMeta: imgMeta,
		})
	}

	err := infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		// 插入图片基础内容
		var errTx error
		noteId, errTx = infra.Dao().NoteDao.Insert(ctx, newNote)
		if errTx != nil {
			return xerror.Wrapf(errTx, "note dao insert tx failed")
		}

		// 填充noteId
		for _, a := range noteAssets {
			a.NoteId = noteId
		}

		// 插入笔记资源数据
		errTx = infra.Dao().NoteAssetRepo.BatchInsert(ctx, noteAssets)
		if errTx != nil {
			return xerror.Wrapf(errTx, "note asset dao batch insert tx failed")
		}

		return nil
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "biz create note failed").WithExtra("note", note).WithCtx(ctx)
	}

	return noteId, nil
}

func (b *NoteCreatorBiz) UpdateNote(ctx context.Context, note *model.UpdateNoteRequest) error {
	var (
		uid = metadata.Uid(ctx)
	)

	now := time.Now().Unix()
	noteId := note.NoteId
	queried, err := infra.Dao().NoteDao.FindOne(ctx, noteId)
	if errors.Is(err, xsql.ErrNoRecord) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "biz find one note failed").WithExtra("note", note).WithCtx(ctx)
	}

	// 确保更新者uid和笔记作者uid相同
	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	newNote := &notedao.Note{
		Id:       noteId,
		Title:    note.Basic.Title,
		Desc:     note.Basic.Desc,
		Privacy:  int8(note.Basic.Privacy),
		Owner:    queried.Owner,
		CreateAt: queried.CreateAt,
		UpdateAt: now,
	}

	assets := make([]*notedao.Asset, 0, len(note.Images))
	for _, img := range note.Images {
		assets = append(assets, &notedao.Asset{
			AssetKey:  img.FileId,
			AssetType: global.AssetTypeImage,
			NoteId:    noteId,
			CreateAt:  now,
		})
	}

	// begin tx
	err = infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		// 先更新基础信息
		err := infra.Dao().NoteDao.Update(ctx, newNote)
		if err != nil {
			return xerror.Wrapf(err, "note dao update tx failed")
		}

		// 找出旧资源
		oldAssets, err := infra.Dao().NoteAssetRepo.FindImageByNoteId(ctx, newNote.Id)
		if err != nil && !errors.Is(err, xsql.ErrNoRecord) {
			return xerror.Wrapf(err, "noteasset dao find failed")
		}

		// 笔记的新资源
		newAssetKeys := make([]string, 0, len(assets))
		for _, asset := range assets {
			newAssetKeys = append(newAssetKeys, asset.AssetKey)
		}

		// 随后删除旧资源
		// 删除除了newAssetKeys之外的其它
		err = infra.Dao().NoteAssetRepo.ExcludeDeleteImageByNoteId(ctx, newNote.Id, newAssetKeys)
		if err != nil {
			return xerror.Wrapf(err, "noteasset dao delete tx failed")
		}

		// 找出old和new的资源差异，只更新发生了变化的部分
		oldAssetMap := make(map[string]struct{})
		for _, old := range oldAssets {
			oldAssetMap[old.AssetKey] = struct{}{}
		}
		newAssets := make([]*notedao.Asset, 0, len(assets))
		for _, asset := range assets {
			if _, ok := oldAssetMap[asset.AssetKey]; !ok {
				newAssets = append(newAssets, &notedao.Asset{
					AssetKey:  asset.AssetKey,
					AssetType: global.AssetTypeImage,
					NoteId:    newNote.Id,
					CreateAt:  now,
				})
			}
		}

		if len(newAssets) == 0 {
			return nil
		}

		// 插入新的资源
		err = infra.Dao().NoteAssetRepo.BatchInsert(ctx, newAssets)
		if err != nil {
			return xerror.Wrapf(err, "noteasset dao batch insert tx failed")
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "biz update note failed").WithExtras("req", note).WithCtx(ctx)
	}

	return nil
}

func (b *NoteCreatorBiz) DeleteNote(ctx context.Context, note *model.DeleteNoteRequest) error {
	var (
		uid    int64 = metadata.Uid(ctx)
		noteId       = note.NoteId
	)

	queried, err := infra.Dao().NoteDao.FindOne(ctx, noteId)
	if errors.Is(err, xsql.ErrNoRecord) {
		return global.ErrNoteNotFound
	}
	if err != nil {
		return xerror.Wrapf(err, "repo find one note failed").WithExtra("req", note).WithCtx(ctx)
	}

	if uid != queried.Owner {
		return global.ErrPermDenied.Msg("你不拥有该笔记")
	}

	err = infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		err := infra.Dao().NoteDao.Delete(ctx, noteId)
		if err != nil {
			return xerror.Wrapf(err, "dao delete note basic tx failed")
		}

		err = infra.Dao().NoteAssetRepo.DeleteByNoteId(ctx, noteId)
		if err != nil {
			return xerror.Wrapf(err, "dao delete note asset tx failed")
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "biz delete note failed").WithExtras("req", note).WithCtx(ctx)
	}

	return nil
}

func (b *NoteCreatorBiz) CreatorGetNote(ctx context.Context, noteId int64) (*model.Note, error) {
	var (
		uid = metadata.Uid(ctx)
		nid = noteId
	)

	note, err := infra.Dao().NoteDao.FindOne(ctx, nid)
	if err != nil {
		if xsql.IsNotFound(err) {
			return nil, global.ErrNoteNotFound
		}
		return nil, xerror.Wrapf(err, "biz get note failed")
	}

	if uid != note.Owner {
		return nil, global.ErrNotNoteOwner
	}

	res, err := b.AssembleNotes(ctx, model.NoteFromDao(note).AsSlice())
	if err != nil || len(res.Items) == 0 {
		return nil, xerror.Wrapf(err, "assemble notes failed").WithExtra("noteId", noteId).WithCtx(ctx)
	}

	return res.Items[0], nil
}

func (b *NoteCreatorBiz) ListNote(ctx context.Context) (*model.Notes, error) {
	var (
		uid = metadata.Uid(ctx)
	)

	notes, err := infra.Dao().NoteDao.ListByOwner(ctx, uid)
	if errors.Is(err, xsql.ErrNoRecord) {
		return &model.Notes{}, nil
	}
	if err != nil {
		return nil, xerror.Wrapf(err, "biz note list by owner failed").WithCtx(ctx)
	}

	return b.AssembleNotes(ctx, model.NoteSliceFromDao(notes))
}

func (b *NoteCreatorBiz) PageListNoteWithCursor(ctx context.Context, cursor int64, count int32) (*model.Notes, model.PageResult, error) {
	var (
		uid      = metadata.Uid(ctx)
		nextPage = model.PageResult{}
	)

	if cursor == 0 {
		cursor = math.MaxInt64
	}
	notes, err := infra.Dao().NoteDao.ListByOwnerByCursor(ctx, uid, cursor, count)
	if errors.Is(err, xsql.ErrNoRecord) {
		return &model.Notes{}, nextPage, nil
	}
	if err != nil {
		return nil, nextPage,
			xerror.Wrapf(err, "biz note list by owner with cursor failed").
				WithCtx(ctx).
				WithExtras("cursor", cursor, "count", count)
	}

	// 计算下一次请求的游标位置
	if len(notes) > 0 {
		nextPage.NextCursor = notes[len(notes)-1].Id
		if len(notes) == int(count) {
			nextPage.HasNext = true
		}
	}

	notesResp, err := b.AssembleNotes(ctx, model.NoteSliceFromDao(notes))
	if err != nil {
		return nil,
			nextPage,
			xerror.Wrapf(err, "biz note failed to assemble notes when cursor page list notes").WithCtx(ctx).
				WithExtras("cursor", cursor, "count", count)
	}

	return notesResp, nextPage, nil
}

// page从1开始
func (b *NoteCreatorBiz) PageListNote(ctx context.Context, page, count int32) (*model.Notes, int64, error) {
	var (
		uid = metadata.Uid(ctx)
	)

	total, err := infra.Dao().NoteDao.GetPostedCountByOwner(ctx, uid)
	if err != nil {
		if !errors.Is(err, xsql.ErrNoRecord) {
			return nil, 0, xerror.Wrapf(err, "biz note count by owner failed").WithCtx(ctx)
		}

		return &model.Notes{}, 0, nil
	}

	notes, err := infra.Dao().NoteDao.PageListByOwner(ctx, uid, page, count)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "biz note page list failed").WithCtx(ctx)
	}

	notesResp, err := b.AssembleNotes(ctx, model.NoteSliceFromDao(notes))
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "biz note failed to assemble notes when page list notes")
	}

	return notesResp, total, nil
}

func (b *NoteCreatorBiz) GetUploadAuth(ctx context.Context, req *model.UploadAuthRequest) (*model.UploadAuthResponse, error) {
	return nil, xerror.Wrap(global.ErrPermDenied.Msg("服务器签名失败"))
}

func (b *NoteCreatorBiz) GetUploadAuthSTS(ctx context.Context,
	req *model.UploadAuthRequest) (*model.UploadAuthSTSResponse, error) {
	// 生成count个上传凭证
	fileIds := make([]string, 0, req.Count)
	for range req.Count {
		fileIds = append(fileIds, b.OssKeyGen.Gen())
	}

	now := time.Now()
	expireAt := now.Add(config.Conf.UploadAuthSign.JwtDuration)
	claim := newStsUploadAuthClaim(now, expireAt, req.Resource, req.Source)

	token := jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, claim)
	ss, err := token.SignedString(utils.StringToBytes(config.Conf.UploadAuthSign.Sk))
	if err != nil {
		return nil, xerror.Wrapf(global.ErrInternal.Msg("服务器签名失败"), "%s", err.Error()).
			WithCtx(ctx)
	}

	return &model.UploadAuthSTSResponse{
		FileIds:     fileIds,
		CurrentTime: now.Unix(),
		ExpireTime:  expireAt.Unix(),
		UploadAddr:  config.Conf.Oss.UploadEndpoint,
		Token:       ss,
	}, nil
}

type stsUploadAuthClaim struct {
	jwtv5.RegisteredClaims

	AccessKey string `json:"access_key"`
	Resource  string `json:"resource"`
	Source    string `json:"source"`
}

var claimSk = []byte(config.Conf.UploadAuthSign.Sk)

func newStsUploadAuthClaim(now, expireAt time.Time, resource, source string) *stsUploadAuthClaim {
	akb := make([]byte, 16)
	_, err := rand.Read(akb)
	if err != nil {
		akb = []byte(config.Conf.UploadAuthSign.Ak)
	}

	hash := hmac.New(sha1.New, claimSk)
	hash.Write(akb)
	akb = hash.Sum(nil)
	ak := hex.EncodeToString(akb)

	return &stsUploadAuthClaim{
		AccessKey: ak,
		Resource:  resource,
		Source:    source,

		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    config.Conf.UploadAuthSign.JwtIssuer,
			Subject:   config.Conf.UploadAuthSign.JwtSubject,
			ID:        config.Conf.UploadAuthSign.JwtId,
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
			ExpiresAt: jwtv5.NewNumericDate(expireAt),
		},
	}
}
