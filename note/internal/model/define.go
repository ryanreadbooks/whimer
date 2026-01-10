package model

import (
	"slices"

	v1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type Privacy int8

// 笔记可见范围
const (
	// 	公开
	PrivacyPublic = Privacy(v1.NotePrivacy_PUBLIC)

	// 私有
	PrivacyPrivate = Privacy(v1.NotePrivacy_PRIVATE)
)

type NoteType int8

func (t NoteType) String() string {
	switch t {
	case NoteTypeImage:
		return "image"
	case NoteTypeVideo:
		return "video"
	default:
		return ""
	}
}

type AssetType = NoteType

// 笔记资源类型
const (
	NoteTypeImage = NoteType(v1.NoteAssetType_IMAGE)

	NoteTypeVideo = NoteType(v1.NoteAssetType_VIDEO)
)

const (
	// 图片
	AssetTypeImage = NoteType(v1.NoteAssetType_IMAGE)

	// 视频
	AssetTypeVideo = NoteType(v1.NoteAssetType_VIDEO)
)

type NoteState int8

// 笔记状态流转:
//
//	Init ──▶ Processing ──▶ Processed ──▶ Auditing ──▶ AuditPassed ──▶ Published
//	              │                           │                            │
//	           处理失败                    审核不通过                       违规
//	             ▼                           ▼                           ▼
//	        ProcessFailed                 Rejected                      Banned
//
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

// 对外暴露状态
func NoteStateAsLifeCycleState(state NoteState) v1.NoteLifeCycleState {
	switch state {
	case NoteStatePublished:
		return v1.NoteLifeCycleState_LIFE_CYCLE_STATE_PUBLISHED
	case NoteStateBanned:
		return v1.NoteLifeCycleState_LIFE_CYCLE_STATE_BANNED
	case NoteStateProcessFailed, NoteStateRejected:
		return v1.NoteLifeCycleState_LIFE_CYCLE_STATE_REJECTED
	default:
		return v1.NoteLifeCycleState_LIFE_CYCLE_STATE_AUDITING
	}
}

func LifeCycleNotePublished() []NoteState {
	return []NoteState{NoteStatePublished}
}

func LifeCycleNoteAuditing() []NoteState {
	return []NoteState{
		NoteStateInit,
		NoteStateAuditing,
		NoteStateProcessing,
		NoteStateProcessed,
		NoteStateAuditPassed,
	}
}

func LifeCycleNoteRejected() []NoteState {
	return []NoteState{NoteStateRejected, NoteStateProcessFailed}
}

func LifeCycleNoteBanned() []NoteState {
	return []NoteState{NoteStateBanned}
}

func IsNoteStateConsideredAsPublished(state NoteState) bool {
	return slices.Contains(LifeCycleNotePublished(), state)
}

func IsNoteStateConsideredAsAuditing(state NoteState) bool {
	return slices.Contains(LifeCycleNoteAuditing(), state)
}

func IsNoteStateConsideredAsRejected(state NoteState) bool {
	return slices.Contains(LifeCycleNoteRejected(), state)
}

func IsNoteStateConsideredAsBanned(state NoteState) bool {
	return state == NoteStateBanned
}

type ProcedureStatus int8

// 本地流程处理状态记录
const (
	// 处理中
	ProcedureStatusProcessing ProcedureStatus = 0

	// 处理成功
	ProcedureStatusSuccess ProcedureStatus = 1

	// 处理失败
	ProcedureStatusFailed ProcedureStatus = 2
)

type ProcedureType string

const (
	// 资源处理流程
	ProcedureTypeAssetProcess ProcedureType = "asset_process"

	// 审核流程（预留）
	ProcedureTypeAudit ProcedureType = "audit"

	// 发布流程
	ProcedureTypePublish ProcedureType = "publish"
)

// 流程对应笔记状态
//
// 用以判断当前笔记流程进行到哪一步 用于中断正在进行的流程
func MapNoteStateToProcedureType(state NoteState) ProcedureType {
	switch state {
	case NoteStateInit, NoteStateProcessing, NoteStateProcessFailed:
		return ProcedureTypeAssetProcess
	case NoteStateProcessed, NoteStateAuditing, NoteStateRejected:
		return ProcedureTypeAudit
	case NoteStateAuditPassed, NoteStatePublished, NoteStateBanned:
		return ProcedureTypePublish
	}

	return ""
}
