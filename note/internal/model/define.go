package model

import v1 "github.com/ryanreadbooks/whimer/note/api/v1"

type Privacy int8

// 笔记可见范围
const (
	// 	公开
	PrivacyPublic = Privacy(v1.NotePrivacy_PUBLIC)

	// 私有
	PrivacyPrivate = Privacy(v1.NotePrivacy_PRIVATE)
)

type NoteType int8

type AssetType = NoteType

// 笔记资源类型
const (
	// 图片
	AssetTypeImage = NoteType(v1.NoteAssetType_IMAGE)

	// 视频
	AssetTypeVideo = NoteType(v1.NoteAssetType_VIDEO)
)

type NoteState int8

// 笔记状态
const (
	// 初始状态
	NoteStateInit = NoteState(v1.NoteState_NOTE_STATE_UNSPECIFIED)

	// 资源处理中
	NoteStateProcessing = NoteState(v1.NoteState_PROCESSING)

	// 资源处理完成
	NoteStateProcessed = NoteState(v1.NoteState_PROCESSED)

	// 资源处理失败
	NoteStateProcessFailed = NoteState(v1.NoteState_PROCESS_FAILED)

	// 审核中
	NoteStateAuditing = NoteState(v1.NoteState_AUDITING)

	// 审核不通过
	NoteStateRejected = NoteState(v1.NoteState_REJECTED)

	// 审核通过
	NoteStateAuditPassed = NoteState(v1.NoteState_AUDIT_PASSED)

	// 已发布
	NoteStatePublished = NoteState(v1.NoteState_PUBLISHED)

	// 被封禁
	NoteStateBanned = NoteState(v1.NoteState_BANNED)
)

type ProcedureStatus int8

// 本地流程处理状态记录
const (
	// 处理中
	ProcessStatusProcessing ProcedureStatus = 0

	// 处理成功
	ProcessStatusSuccess ProcedureStatus = 1

	// 处理失败
	ProcessStatusFailed ProcedureStatus = 2
)

type ProcedureType string

const (
	ProcedureTypeAssetProcess ProcedureType = "asset_process"
)
