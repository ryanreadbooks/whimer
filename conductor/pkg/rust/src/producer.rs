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

pub type Result<T> = std::result::Result<T, ClientError>;

#[derive(Debug)]
pub enum ClientError {
  Discovery(DiscoveryError),
  GrpcStatus(tonic::Status),
  Json(serde_json::Error),
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

#[derive(Debug, Clone)]
pub struct ClientOptions {
  pub hosts: Vec<String>, // etcd hosts
  pub host_key: String,   // service key
  pub namespace: String,
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

#[derive(Debug, Clone)]
pub struct Task {
  pub id: String,
  pub namespace: String,
  pub task_type: String,
  pub input_args: Vec<u8>,
  pub output_args: Vec<u8>,
  pub callback_url: String,
  pub state: String,
  pub max_retry_cnt: i64,
  pub expire_time: i64,
  pub ctime: i64,
  pub utime: i64,
  pub trace_id: String,
}

impl Task {
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

#[derive(Debug, Clone, Default)]
pub struct ScheduleOptions {
  pub namespace: Option<String>,
  pub callback_url: String,
  pub max_retry: i64,
  pub expire_time: Option<SystemTime>,
  pub expire_after: Option<Duration>,
}

#[derive(Debug, Clone)]
pub struct Client {
  opts: ClientOptions,
  client: TaskServiceClient<tonic::transport::Channel>,
}

impl Client {
  pub async fn new(opts: ClientOptions) -> Result<Self> {
    let etcd_conf = EtcdConfig::new(opts.hosts.clone(), opts.host_key.clone());
    let discovery = EtcdDiscovery::new(etcd_conf).await?;
    let channel = discovery
      .balanced_channel(opts.balance_channel_capacity)
      .await?;
    let client = TaskServiceClient::new(channel);

    Ok(Self { opts, client })
  }

  // 提交任务后返回 通过设置callback接收任务执行结果
  //
  // 返回任务id可用后续查询任务状态
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

  // 获取任务信息
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

  // 终止任务
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
