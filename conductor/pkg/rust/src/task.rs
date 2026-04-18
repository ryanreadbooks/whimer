//! 任务相关的通用类型定义。

/// 任务执行状态。
pub enum TaskState {
  /// 任务执行中。
  Running,
  /// 任务执行成功。
  Success,
  /// 任务执行失败。
  Failure,
}
