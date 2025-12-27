package procedure

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// pipelineStage 流水线阶段
type pipelineStage = model.ProcedureType

// 流水线开始位置
func StartAtAssetProcess() pipelineStage {
	return model.ProcedureTypeAssetProcess
}

func StartAtAudit() pipelineStage {
	return model.ProcedureTypeAudit
}

func StartAtPublish() pipelineStage {
	return model.ProcedureTypePublish
}

var (
	ErrPipelineJobDuplicate = fmt.Errorf("pipeline job duplicate")
	ErrPipelineJobEmpty     = fmt.Errorf("pipeline job empty")
)

// 笔记发布流程流水线
type pipeline struct {
	mgr *Manager

	// 记录当前流程的下一个流程
	nextProcMap map[model.ProcedureType]model.ProcedureType

	// 流程顺序
	procSeqs []model.ProcedureType
}

func (p *pipeline) first() model.ProcedureType {
	return p.procSeqs[0]
}

func (p *pipeline) startAt(stage pipelineStage) model.ProcedureType {
	return stage
}

func (p *pipeline) nextOf(procType model.ProcedureType) model.ProcedureType {
	return p.nextProcMap[procType]
}

// 组装流水线
type pipelineAssembler struct {
	mgr *Manager

	// 按顺序执行的流程
	procSeqs []model.ProcedureType
}

func newPipelineAssembler(mgr *Manager) *pipelineAssembler {
	b := &pipelineAssembler{mgr: mgr}
	return b
}

func (b *pipelineAssembler) addProcedure(procType model.ProcedureType) *pipelineAssembler {
	b.procSeqs = append(b.procSeqs, procType)
	return b
}

func (b *pipelineAssembler) assemble() (*pipeline, error) {
	// 不允许重复
	tmp := xslice.Uniq(b.procSeqs)
	if len(tmp) != len(b.procSeqs) {
		return nil, ErrPipelineJobDuplicate
	}

	if len(b.procSeqs) == 0 {
		return nil, ErrPipelineJobEmpty
	}

	p := &pipeline{
		mgr:         b.mgr,
		procSeqs:    b.procSeqs,
		nextProcMap: make(map[model.ProcedureType]model.ProcedureType),
	}

	// 需要保证所有procType都已经注册 防止外面配置错误
	for _, procType := range b.procSeqs {
		_, ok := b.mgr.registry.Get(procType)
		if !ok {
			return nil, ErrProcedureNotRegistered
		}
	}

	for i := 0; i < len(b.procSeqs)-1; i++ {
		p.nextProcMap[b.procSeqs[i]] = b.procSeqs[i+1]
	}

	xlog.Msgf("pipeline assembled, %v, %v", b.procSeqs, p.nextProcMap).Info()

	return p, nil
}

// 定义内置流水线

// 标准流水线
//
// 标准笔记发布流程: 资源处理 -> 审核 -> 发布
func innerStandardPipeline(mgr *Manager) (*pipeline, error) {
	assembler := newPipelineAssembler(mgr)
	return assembler.
		addProcedure(model.ProcedureTypeAssetProcess).
		addProcedure(model.ProcedureTypeAudit). // TODO 审核暂未实现
		addProcedure(model.ProcedureTypePublish).
		assemble()
}
