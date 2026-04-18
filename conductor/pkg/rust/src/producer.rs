//! 任务生产端（Producer）SDK。
//!
//! 该模块用于向 Conductor 注册任务，并提供任务查询与主动终止能力。

use misc_rs::xgrpc::discovery::{DiscoveryError, EtcdConfig, EtcdDiscovery};
use prost_types::Timestamp;
use serde::Serialize;
use serde::de::DeserializeOwned;
use std::fmt::{Display, Formatter};
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use whimer_idl_rust::conductor::api::task::v1 as taskv1;
use whimer_idl_rust::conductor::api::taskservice::v1 as taskservicev1;
use whimer_idl_rust::conductor::api::taskservice::v1::task_service_client::TaskServiceClient;

const DEFAULT_BALANCE_CHANNEL_CAPACITY: usize = 64;

/// Producer 侧统一返回类型。
pub type Result<T> = std::result::Result<T, ClientError>;

/// Producer 客户端错误类型。
#[derive(Debug)]
pub enum ClientError {
  /// 服务发现阶段错误（例如 etcd 连接或订阅失败）。
  Discovery(DiscoveryError),
  /// gRPC 调用返回错误状态。
  GrpcStatus(tonic::Status),
  /// JSON 序列化或反序列化错误。
  Json(serde_json::Error),
  /// 系统时间转换为 protobuf 时间戳时失败。
  InvalidSystemTime(String),
}

impl Display for ClientError {
  fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
    match self {
      Self::Discovery(err) => write!(f, "discovery error: {err}"),
      Self::GrpcStatus(err) => write!(f, "grpc status: {err}"),
      Self::Json(err) => write!(f, "json error: {err}"),
      Self::InvalidSystemTime(err) => write!(f, "invalid system time: {err}"),
    }
  }
}

impl std::error::Error for ClientError {}

impl From<DiscoveryError> for ClientError {
  fn from(value: DiscoveryError) -> Self {
    Self::Discovery(value)
  }
}

impl From<tonic::Status> for ClientError {
  fn from(value: tonic::Status) -> Self {
    Self::GrpcStatus(value)
  }
}

impl From<serde_json::Error> for ClientError {
  fn from(value: serde_json::Error) -> Self {
    Self::Json(value)
  }
}

/// Producer 客户端初始化配置。
#[derive(Debug, Clone)]
pub struct ClientOptions {
  /// etcd 地址列表，例如 `["localhost:2379"]`。
  pub hosts: Vec<String>,
  /// 服务发现 key，例如 `whimer.conductor.rpc`。
  pub host_key: String,
  /// 默认业务命名空间。
  pub namespace: String,
  /// gRPC 负载均衡通道容量。
  pub balance_channel_capacity: usize,
}

impl Default for ClientOptions {
  fn default() -> Self {
    Self {
      hosts: Vec::new(),
      host_key: String::new(),
      namespace: String::new(),
      balance_channel_capacity: DEFAULT_BALANCE_CHANNEL_CAPACITY,
    }
  }
}

impl ClientOptions {
  /// 创建 Producer 配置。
  ///
  /// `namespace` 会作为未显式传参时的默认命名空间。
  pub fn new(
    hosts: Vec<String>,
    host_key: impl Into<String>,
    namespace: impl Into<String>,
  ) -> Self {
    Self {
      hosts,
      host_key: host_key.into(),
      namespace: namespace.into(),
      balance_channel_capacity: DEFAULT_BALANCE_CHANNEL_CAPACITY,
    }
  }
}

/// 任务信息（Producer 侧视图）。
#[derive(Debug, Clone)]
pub struct Task {
  /// 任务 ID。
  pub id: String,
  /// 所属命名空间。
  pub namespace: String,
  /// 任务类型。
  pub task_type: String,
  /// 输入参数（JSON bytes）。
  pub input_args: Vec<u8>,
  /// 输出参数（JSON bytes）。
  pub output_args: Vec<u8>,
  /// 回调地址。
  pub callback_url: String,
  /// 任务状态（字符串形式）。
  pub state: String,
  /// 最大重试次数。
  pub max_retry_cnt: i64,
  /// 过期时间（Unix ms）。
  pub expire_time: i64,
  /// 创建时间（Unix ms）。
  pub ctime: i64,
  /// 更新时间（Unix ms）。
  pub utime: i64,
  /// 链路追踪 ID。
  pub trace_id: String,
}

impl Task {
  /// 将 `output_args` 反序列化为指定类型。
  ///
  /// 当任务输出为空时返回 `Ok(None)`。
  pub fn unmarshal_output<T: DeserializeOwned>(
    &self,
  ) -> std::result::Result<Option<T>, serde_json::Error> {
    if self.output_args.is_empty() {
      return Ok(None);
    }

    let val = serde_json::from_slice::<T>(&self.output_args)?;
    Ok(Some(val))
  }
}

/// 任务注册选项。
#[derive(Debug, Clone, Default)]
pub struct ScheduleOptions {
  /// 覆盖默认命名空间；为空时回退到客户端默认值。
  pub namespace: Option<String>,
  /// 任务回调 URL。
  pub callback_url: String,
  /// 最大重试次数。
  pub max_retry: i64,
  /// 绝对过期时间。
  pub expire_time: Option<SystemTime>,
  /// 相对过期时间（基于当前时间偏移）。
  pub expire_after: Option<Duration>,
}

/// Producer 客户端。
#[derive(Debug, Clone)]
pub struct Client {
  opts: ClientOptions,
  client: TaskServiceClient<tonic::transport::Channel>,
}

impl Client {
  /// 创建 Producer 客户端。
  ///
  /// 会通过 etcd 发现 Conductor 服务并建立带负载均衡的 gRPC 通道。
  pub async fn new(opts: ClientOptions) -> Result<Self> {
    let etcd_conf = EtcdConfig::new(opts.hosts.clone(), opts.host_key.clone());
    let discovery = EtcdDiscovery::new(etcd_conf).await?;
    let channel = discovery
      .balanced_channel(opts.balance_channel_capacity)
      .await?;
    let client = TaskServiceClient::new(channel);

    Ok(Self { opts, client })
  }

  /// 提交任务。
  ///
  /// 成功后返回任务 ID，可用于后续查询任务状态或主动终止任务。
  pub async fn schedule<T: Serialize>(
    &self,
    task_type: impl Into<String>,
    input: Option<&T>,
    opts: ScheduleOptions,
  ) -> Result<String> {
    let input_args = if let Some(input) = input {
      serde_json::to_vec(input)?
    } else {
      Vec::new()
    };

    self.schedule_raw(task_type, input_args, opts).await
  }

  async fn schedule_raw(
    &self,
    task_type: impl Into<String>,
    input_args: Vec<u8>,
    opts: ScheduleOptions,
  ) -> Result<String> {
    let namespace = opts
      .namespace
      .filter(|namespace| !namespace.is_empty())
      .unwrap_or_else(|| self.opts.namespace.clone());

    let mut req = taskservicev1::RegisterTaskRequest {
      task_type: task_type.into(),
      namespace,
      input_args: input_args.into(),
      callback_url: opts.callback_url,
      max_retry_cnt: opts.max_retry,
      expire_time: None,
    };

    if let Some(expire_time) = opts.expire_time {
      req.expire_time = Some(system_time_to_timestamp(expire_time)?);
    } else if let Some(expire_after) = opts.expire_after {
      req.expire_time = Some(system_time_to_timestamp(SystemTime::now() + expire_after)?);
    }

    let mut client = self.client.clone();
    let resp = client.register_task(req).await?.into_inner();
    Ok(resp.task_id)
  }

  /// 查询任务详情。
  pub async fn get_task(&self, task_id: impl Into<String>) -> Result<Option<Task>> {
    let mut client = self.client.clone();
    let resp = client
      .get_task(taskservicev1::GetTaskRequest {
        task_id: task_id.into(),
      })
      .await?
      .into_inner();

    Ok(resp.task.map(task_from_proto))
  }

  /// 主动终止任务。
  pub async fn abort_task(&self, task_id: impl Into<String>) -> Result<()> {
    let mut client = self.client.clone();
    client
      .abort_task(taskservicev1::AbortTaskRequest {
        task_id: task_id.into(),
      })
      .await?;

    Ok(())
  }
}

fn task_from_proto(task: taskv1::Task) -> Task {
  Task {
    id: task.id,
    namespace: task.namespace,
    task_type: task.task_type,
    input_args: task.input_args.to_vec(),
    output_args: task
      .output_args
      .map(|args| args.to_vec())
      .unwrap_or_default(),
    callback_url: task.callback_url,
    state: task.state,
    max_retry_cnt: task.max_retry_cnt,
    expire_time: task.expire_time,
    ctime: task.ctime,
    utime: task.utime,
    trace_id: task.trace_id,
  }
}

fn system_time_to_timestamp(system_time: SystemTime) -> Result<Timestamp> {
  let duration = system_time
    .duration_since(UNIX_EPOCH)
    .map_err(|err| ClientError::InvalidSystemTime(err.to_string()))?;

  Ok(Timestamp {
    seconds: duration.as_secs() as i64,
    nanos: duration.subsec_nanos() as i32,
  })
}
