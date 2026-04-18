//! Conductor Rust SDK。
//!
//! 当前 crate 提供三类能力：
//! - `producer`：任务生产端，负责注册/查询/终止任务。
//! - `worker`：任务消费端，负责拉取任务并执行处理逻辑。
//! - `task`：通用任务状态定义。

pub mod producer;
pub mod task;
pub mod worker;
