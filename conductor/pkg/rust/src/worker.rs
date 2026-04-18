//! 任务消费端（Worker）SDK。
//!
//! 该模块负责从 Conductor 拉取任务、执行用户处理函数、
//! 定期上报心跳/进度，并在任务完成后回传执行结果。

use misc_rs::xgrpc::discovery::{DiscoveryError, EtcdConfig, EtcdDiscovery};
use serde::Serialize;
use serde::de::DeserializeOwned;
use std::collections::{HashMap, HashSet};
use std::fmt::{Display, Formatter};
use std::future::Future;
use std::pin::Pin;
use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::{Arc, RwLock};
use std::time::Duration;
use tokio::sync::{Notify, OwnedSemaphorePermit, Semaphore, watch};
use tokio::time::sleep;
use tonic::Code;
use whimer_idl_rust::conductor::api::task::v1 as taskv1;
use whimer_idl_rust::conductor::api::worker::v1 as workerv1;
use whimer_idl_rust::conductor::api::workerservice::v1 as workerservicev1;
use whimer_idl_rust::conductor::api::workerservice::v1::worker_service_client::WorkerServiceClient;

const DEFAULT_CONCURRENCY: usize = 1;
const DEFAULT_BALANCE_CHANNEL_CAPACITY: usize = 64;
const DEFAULT_HEARTBEAT_INTERVAL: Duration = Duration::from_secs(10);
const DEFAULT_REPORT_RETRY_MAX_ATTEMPTS: usize = 5;
const DEFAULT_REPORT_RETRY_INITIAL_BACKOFF: Duration = Duration::from_millis(100);
const DEFAULT_REPORT_RETRY_MAX_BACKOFF: Duration = Duration::from_secs(10);
const DEFAULT_REPORT_RETRY_MULTIPLIER: f64 = 2.0;
const POLL_ERROR_BACKOFF: Duration = Duration::from_secs(1);

/// Worker 侧统一返回类型。
pub type Result<T> = std::result::Result<T, WorkerError>;

/// Worker 客户端错误类型。
#[derive(Debug)]
pub enum WorkerError {
  /// 服务发现阶段错误（例如 etcd 连接失败）。
  Discovery(DiscoveryError),
  /// gRPC 调用返回错误状态。
  GrpcStatus(tonic::Status),
  /// JSON 编解码错误。
  Json(serde_json::Error),
}

impl Display for WorkerError {
  fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
    match self {
      Self::Discovery(err) => write!(f, "discovery error: {err}"),
      Self::GrpcStatus(err) => write!(f, "grpc status: {err}"),
      Self::Json(err) => write!(f, "json error: {err}"),
    }
  }
}

impl std::error::Error for WorkerError {}

impl From<DiscoveryError> for WorkerError {
  fn from(value: DiscoveryError) -> Self {
    Self::Discovery(value)
  }
}

impl From<tonic::Status> for WorkerError {
  fn from(value: tonic::Status) -> Self {
    Self::GrpcStatus(value)
  }
}

impl From<serde_json::Error> for WorkerError {
  fn from(value: serde_json::Error) -> Self {
    Self::Json(value)
  }
}

/// 任务完成上报的重试策略。
#[derive(Debug, Clone)]
pub struct RetryOptions {
  /// 最大重试次数。
  pub max_attempts: usize,
  /// 初始退避时间。
  pub initial_backoff: Duration,
  /// 最大退避时间。
  pub max_backoff: Duration,
  /// 退避倍率（指数退避）。
  pub multiplier: f64,
}

impl Default for RetryOptions {
  fn default() -> Self {
    Self {
      max_attempts: DEFAULT_REPORT_RETRY_MAX_ATTEMPTS,
      initial_backoff: DEFAULT_REPORT_RETRY_INITIAL_BACKOFF,
      max_backoff: DEFAULT_REPORT_RETRY_MAX_BACKOFF,
      multiplier: DEFAULT_REPORT_RETRY_MULTIPLIER,
    }
  }
}

/// Worker 初始化配置。
#[derive(Debug, Clone)]
pub struct WorkerOptions {
  /// etcd 地址列表。
  pub hosts: Vec<String>,
  /// 服务发现 key。
  pub host_key: String,
  /// Worker 唯一标识。
  pub worker_id: String,
  /// Worker 实例 IP（用于上报元信息）。
  pub ip: String,
  /// 并发处理上限。
  pub concurrency: usize,
  /// 结果上报重试策略。
  pub report_retry: RetryOptions,
  /// 心跳上报间隔。
  pub heartbeat_interval: Duration,
  /// gRPC 负载均衡通道容量。
  pub balance_channel_capacity: usize,
}

impl Default for WorkerOptions {
  fn default() -> Self {
    Self {
      hosts: Vec::new(),
      host_key: String::new(),
      worker_id: String::new(),
      ip: String::new(),
      concurrency: DEFAULT_CONCURRENCY,
      report_retry: RetryOptions::default(),
      heartbeat_interval: DEFAULT_HEARTBEAT_INTERVAL,
      balance_channel_capacity: DEFAULT_BALANCE_CHANNEL_CAPACITY,
    }
  }
}

impl WorkerOptions {
  /// 创建 Worker 配置。
  pub fn new(hosts: Vec<String>, host_key: impl Into<String>) -> Self {
    Self {
      hosts,
      host_key: host_key.into(),
      ..Self::default()
    }
  }
}

/// Worker 侧任务信息。
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
  /// 回调地址。
  pub callback_url: String,
  /// 最大重试次数。
  pub max_retry_cnt: i64,
  /// 过期时间（Unix ms）。
  pub expire_time: i64,
  /// 创建时间（Unix ms）。
  pub ctime: i64,
  /// 链路追踪 ID。
  pub trace_id: String,
}

impl Task {
  /// 将 `input_args` 反序列化为指定类型。
  ///
  /// 当输入为空时返回 `Ok(None)`。
  pub fn unmarshal_input<T: DeserializeOwned>(
    &self,
  ) -> std::result::Result<Option<T>, serde_json::Error> {
    if self.input_args.is_empty() {
      return Ok(None);
    }
    let val = serde_json::from_slice::<T>(&self.input_args)?;
    Ok(Some(val))
  }
}

/// 任务处理结果。
///
/// 当 `error` 为空时表示执行成功，否则表示失败；
/// `retryable` 控制失败后是否允许由服务端重试。
#[derive(Debug, Clone)]
pub struct HandlerResult {
  /// 输出参数（JSON bytes）。
  pub output_args: Vec<u8>,
  /// 错误信息。
  pub error: Option<String>,
  /// 失败是否可重试。
  pub retryable: bool,
}

impl Default for HandlerResult {
  fn default() -> Self {
    Self::success()
  }
}

impl HandlerResult {
  /// 构造一个成功结果（无输出）。
  pub fn success() -> Self {
    Self {
      output_args: Vec::new(),
      error: None,
      retryable: false,
    }
  }

  /// 构造一个成功结果，并将输出序列化为 JSON。
  pub fn success_json<T: Serialize>(output: &T) -> std::result::Result<Self, serde_json::Error> {
    Ok(Self {
      output_args: serde_json::to_vec(output)?,
      error: None,
      retryable: false,
    })
  }

  /// 构造一个失败结果（不可重试）。
  pub fn failure(err: impl Into<String>) -> Self {
    Self {
      output_args: Vec::new(),
      error: Some(err.into()),
      retryable: false,
    }
  }

  /// 构造一个失败结果（可重试）。
  pub fn retryable_failure(err: impl Into<String>) -> Self {
    Self {
      output_args: Vec::new(),
      error: Some(err.into()),
      retryable: true,
    }
  }
}

/// 进度提供器。
///
/// Worker 心跳上报时会调用该接口读取当前进度。
pub trait ProgressProvider: Send + Sync {
  /// 返回当前进度，建议范围 `0..=100`。
  fn progress(&self) -> i64;
}

impl<F> ProgressProvider for F
where
  F: Fn() -> i64 + Send + Sync,
{
  fn progress(&self) -> i64 {
    (self)()
  }
}

type BoxedHandlerFuture = Pin<Box<dyn Future<Output = HandlerResult> + Send>>;

trait SimpleHandlerFn: Send + Sync {
  fn call(&self, task: Task) -> BoxedHandlerFuture;
}

impl<F, Fut> SimpleHandlerFn for F
where
  F: Fn(Task) -> Fut + Send + Sync + 'static,
  Fut: Future<Output = HandlerResult> + Send + 'static,
{
  fn call(&self, task: Task) -> BoxedHandlerFuture {
    Box::pin((self)(task))
  }
}

trait TaskHandlerFn: Send + Sync {
  fn call(&self, tc: TaskContext) -> BoxedHandlerFuture;
}

impl<F, Fut> TaskHandlerFn for F
where
  F: Fn(TaskContext) -> Fut + Send + Sync + 'static,
  Fut: Future<Output = HandlerResult> + Send + 'static,
{
  fn call(&self, tc: TaskContext) -> BoxedHandlerFuture {
    Box::pin((self)(tc))
  }
}

/// 任务执行上下文。
///
/// 可用于查询任务信息、检测终止信号和上报进度。
#[derive(Clone)]
pub struct TaskContext {
  inner: Arc<TaskContextInner>,
}

struct TaskContextInner {
  task: Task,
  worker: Worker,
  heartbeat_interval: Duration,
  aborted: AtomicBool,
  abort_notify: Notify,
  stop_tx: watch::Sender<bool>,
  progress_provider: RwLock<Option<Arc<dyn ProgressProvider>>>,
}

impl TaskContext {
  fn new(task: Task, worker: Worker, heartbeat_interval: Duration) -> Self {
    let (stop_tx, _) = watch::channel(false);
    Self {
      inner: Arc::new(TaskContextInner {
        task,
        worker,
        heartbeat_interval,
        aborted: AtomicBool::new(false),
        abort_notify: Notify::new(),
        stop_tx,
        progress_provider: RwLock::new(None),
      }),
    }
  }

  /// 获取当前任务信息。
  pub fn task(&self) -> &Task {
    &self.inner.task
  }

  /// 判断任务是否已被服务端标记为终止。
  pub fn is_aborted(&self) -> bool {
    self.inner.aborted.load(Ordering::SeqCst)
  }

  /// 异步等待任务终止信号。
  pub async fn wait_abort(&self) {
    if self.is_aborted() {
      return;
    }
    self.inner.abort_notify.notified().await;
  }

  /// 设置进度提供器。
  ///
  /// 设置后，心跳线程会自动读取该进度并上报。
  pub fn set_progress_provider<P>(&self, provider: P)
  where
    P: ProgressProvider + 'static,
  {
    let mut guard = self
      .inner
      .progress_provider
      .write()
      .expect("progress provider lock poisoned");
    *guard = Some(Arc::new(provider));
  }

  /// 主动上报进度，并返回是否需要终止任务。
  pub async fn report_progress(&self, progress: i64) -> bool {
    if self.is_aborted() {
      return true;
    }

    let aborted = self
      .inner
      .worker
      .report_task_progress(&self.inner.task.id, progress)
      .await;
    if aborted {
      self.mark_aborted();
    }

    aborted
  }

  fn current_progress(&self) -> i64 {
    let guard = self
      .inner
      .progress_provider
      .read()
      .expect("progress provider lock poisoned");
    if let Some(provider) = guard.as_ref() {
      provider.progress()
    } else {
      -1
    }
  }

  fn mark_aborted(&self) {
    if !self.inner.aborted.swap(true, Ordering::SeqCst) {
      self.inner.abort_notify.notify_waiters();
    }
  }

  fn start_heartbeat(&self) -> tokio::task::JoinHandle<()> {
    let tc = self.clone();
    let mut stop_rx = tc.inner.stop_tx.subscribe();
    tokio::spawn(async move {
      // 心跳线程由两类信号驱动：显式 stop 或定时上报。
      loop {
        tokio::select! {
          changed = stop_rx.changed() => {
            if changed.is_err() || *stop_rx.borrow() {
              return;
            }
          }
          _ = sleep(tc.inner.heartbeat_interval) => {
            if tc.is_aborted() {
              return;
            }
            let progress = tc.current_progress();
            if tc.report_progress(progress).await {
              return;
            }
          }
        }
      }
    })
  }

  fn stop(&self) {
    let _ = self.inner.stop_tx.send(true);
  }
}

/// Worker 客户端。
#[derive(Clone)]
pub struct Worker {
  inner: Arc<WorkerInner>,
}

struct WorkerInner {
  opts: WorkerOptions,
  client: WorkerServiceClient<tonic::transport::Channel>,
  handlers: RwLock<HashMap<String, Arc<dyn SimpleHandlerFn>>>,
  task_handlers: RwLock<HashMap<String, Arc<dyn TaskHandlerFn>>>,
  shutdown_tx: watch::Sender<bool>,
  stopping: AtomicBool,
}

impl Worker {
  /// 创建 Worker 客户端。
  ///
  /// 会通过 etcd 发现 Conductor 服务并建立负载均衡 gRPC 通道。
  pub async fn new(mut opts: WorkerOptions) -> Result<Self> {
    normalize_options(&mut opts);

    let etcd_conf = EtcdConfig::new(opts.hosts.clone(), opts.host_key.clone());
    let discovery = EtcdDiscovery::new(etcd_conf).await?;
    let channel = discovery
      .balanced_channel(opts.balance_channel_capacity)
      .await?;

    let client = WorkerServiceClient::new(channel);
    let (shutdown_tx, _) = watch::channel(false);

    Ok(Self {
      inner: Arc::new(WorkerInner {
        opts,
        client,
        handlers: RwLock::new(HashMap::new()),
        task_handlers: RwLock::new(HashMap::new()),
        shutdown_tx,
        stopping: AtomicBool::new(false),
      }),
    })
  }

  /// 注册简单任务处理器。
  ///
  /// 同一个 `task_type` 后注册会覆盖前一次注册。
  pub fn register_handler<F, Fut>(&self, task_type: impl Into<String>, handler: F)
  where
    F: Fn(Task) -> Fut + Send + Sync + 'static,
    Fut: Future<Output = HandlerResult> + Send + 'static,
  {
    self
      .inner
      .handlers
      .write()
      .expect("handler lock poisoned")
      .insert(task_type.into(), Arc::new(handler));
  }

  /// 注册带上下文的任务处理器。
  ///
  /// 可通过 `TaskContext` 使用进度上报与终止检测能力。
  pub fn register_task_handler<F, Fut>(&self, task_type: impl Into<String>, handler: F)
  where
    F: Fn(TaskContext) -> Fut + Send + Sync + 'static,
    Fut: Future<Output = HandlerResult> + Send + 'static,
  {
    self
      .inner
      .task_handlers
      .write()
      .expect("task handler lock poisoned")
      .insert(task_type.into(), Arc::new(handler));
  }

  /// 启动 Worker 主循环。
  ///
  /// 该方法会持续拉取并处理任务，直到收到 `stop()` 信号。
  pub async fn run(&self) -> Result<()> {
    let task_types = self.registered_task_types();
    if task_types.is_empty() {
      return Ok(());
    }

    let semaphore = Arc::new(Semaphore::new(self.inner.opts.concurrency));
    let mut handles = Vec::with_capacity(task_types.len());
    for task_type in task_types {
      let worker = self.clone();
      let sem = Arc::clone(&semaphore);
      let shutdown_rx = self.inner.shutdown_tx.subscribe();
      handles.push(tokio::spawn(async move {
        worker.poll_loop(task_type, sem, shutdown_rx).await;
      }));
    }

    for handle in handles {
      let _ = handle.await;
    }

    Ok(())
  }

  /// 停止 Worker。
  pub fn stop(&self) {
    if !self.inner.stopping.swap(true, Ordering::SeqCst) {
      let _ = self.inner.shutdown_tx.send(true);
    }
  }

  /// 判断 Worker 是否已进入停止流程。
  pub fn is_stopping(&self) -> bool {
    self.inner.stopping.load(Ordering::SeqCst)
  }

  async fn poll_loop(
    self,
    task_type: String,
    sem: Arc<Semaphore>,
    mut shutdown_rx: watch::Receiver<bool>,
  ) {
    loop {
      if *shutdown_rx.borrow() {
        return;
      }

      // 先拿并发令牌，再执行 LongPoll，确保“轮询+处理”总并发受限。
      let permit = tokio::select! {
        changed = shutdown_rx.changed() => {
          if changed.is_err() || *shutdown_rx.borrow() {
            return;
          }
          continue;
        }
        permit = sem.clone().acquire_owned() => {
          match permit {
            Ok(permit) => permit,
            Err(_) => return,
          }
        }
      };

      match self.poll_task(&task_type).await {
        Ok(Some(task)) if !task.id.is_empty() => {
          let worker = self.clone();
          // 处理任务放到独立协程，令牌生命周期与任务处理绑定。
          tokio::spawn(async move {
            worker.process_task(task, permit).await;
          });
        }
        Ok(_) => {
          drop(permit);
        }
        Err(err) => {
          if self.is_stopping() {
            drop(permit);
            return;
          }
          if !is_timeout_status(&err) {
            eprintln!("worker poll task failed: {err}");
            sleep(POLL_ERROR_BACKOFF).await;
          }
          drop(permit);
        }
      }
    }
  }

  async fn poll_task(&self, task_type: &str) -> std::result::Result<Option<Task>, tonic::Status> {
    let mut client = self.inner.client.clone();
    let req = workerservicev1::LongPollRequest {
      worker: Some(workerv1::Worker {
        id: self.inner.opts.worker_id.clone(),
        ability: Some(workerv1::WorkerAbility {
          task_type: task_type.to_string(),
          task_description: String::new(),
        }),
        metadata: Some(workerv1::WorkerMetadata {
          ip: self.inner.opts.ip.clone(),
          cpu_usage: 0.0,
          mem_usage: 0.0,
        }),
        state: workerv1::WorkerState::Ready as i32,
      }),
    };

    let resp = client.long_poll(req).await?.into_inner();
    Ok(resp.task.map(task_from_proto))
  }

  async fn process_task(&self, task: Task, _permit: OwnedSemaphorePermit) {
    if let Err(err) = self.accept_task(&task.id).await {
      eprintln!("worker accept task failed (task={}): {err}", task.id);
      return;
    }

    // 仅在服务端 Accept 成功后开启心跳，避免未接单任务误上报。
    let tc = TaskContext::new(
      task.clone(),
      self.clone(),
      self.inner.opts.heartbeat_interval,
    );
    let heartbeat_handle = tc.start_heartbeat();
    let result = self
      .execute_handler(&task.task_type, &task, tc.clone())
      .await;
    tc.stop();
    let _ = heartbeat_handle.await;

    // 任务已被服务端终止时，不再上报 complete，避免覆盖终态。
    if tc.is_aborted() {
      eprintln!("worker task aborted by conductor (task={})", task.id);
      return;
    }

    self.complete_task_with_retry(task.id.clone(), result).await;
  }

  async fn execute_handler(&self, task_type: &str, task: &Task, tc: TaskContext) -> HandlerResult {
    // 优先使用 TaskHandler（支持进度与终止检测）。
    let task_handler = {
      self
        .inner
        .task_handlers
        .read()
        .expect("task handler lock poisoned")
        .get(task_type)
        .cloned()
    };
    if let Some(task_handler) = task_handler {
      return run_handler_future(task_handler.call(tc), task.id.clone()).await;
    }

    let handler = {
      self
        .inner
        .handlers
        .read()
        .expect("handler lock poisoned")
        .get(task_type)
        .cloned()
    };
    if let Some(handler) = handler {
      return run_handler_future(handler.call(task.clone()), task.id.clone()).await;
    }

    HandlerResult::failure(format!("no handler for task type: {task_type}"))
  }

  async fn accept_task(&self, task_id: &str) -> std::result::Result<(), tonic::Status> {
    let mut client = self.inner.client.clone();
    client
      .accept_task(workerservicev1::AcceptTaskRequest {
        task_id: task_id.to_string(),
      })
      .await?;
    Ok(())
  }

  async fn report_task_progress(&self, task_id: &str, progress: i64) -> bool {
    let mut client = self.inner.client.clone();
    match client
      .report_task(workerservicev1::ReportTaskRequest {
        task_id: task_id.to_string(),
        progress,
      })
      .await
    {
      Ok(resp) => resp.into_inner().aborted,
      Err(err) => {
        // 进度上报失败默认不中断任务，避免短时网络抖动影响业务执行。
        eprintln!("worker report task progress failed (task={task_id}): {err}");
        false
      }
    }
  }

  async fn complete_task_with_retry(&self, task_id: String, result: HandlerResult) {
    let mut backoff = self.inner.opts.report_retry.initial_backoff;
    let max_attempts = self.inner.opts.report_retry.max_attempts;
    let success = result.error.is_none();
    // 固化请求体，重试时复用同一份语义一致的上报数据。
    let req = workerservicev1::CompleteTaskRequest {
      task_id: task_id.clone(),
      output_args: result.output_args.into(),
      success,
      error_msg: result.error.unwrap_or_default().into_bytes().into(),
      retryable: result.retryable,
    };

    for attempt in 1..=max_attempts {
      let mut client = self.inner.client.clone();
      match client.complete_task(req.clone()).await {
        Ok(_) => return,
        Err(err) => {
          if attempt == max_attempts {
            eprintln!(
              "worker complete task failed after max attempts (task={}, attempts={}): {}",
              task_id, attempt, err
            );
            return;
          }

          eprintln!(
            "worker complete task failed, retrying (task={}, attempt={}, max_attempts={}): {}",
            task_id, attempt, max_attempts, err
          );

          // 指数退避 + 上限裁剪，避免重试风暴。
          sleep(backoff).await;
          backoff = next_backoff(backoff, &self.inner.opts.report_retry);
        }
      }
    }
  }

  fn registered_task_types(&self) -> Vec<String> {
    let mut task_types = HashSet::new();

    for task_type in self
      .inner
      .handlers
      .read()
      .expect("handler lock poisoned")
      .keys()
    {
      task_types.insert(task_type.clone());
    }

    for task_type in self
      .inner
      .task_handlers
      .read()
      .expect("task handler lock poisoned")
      .keys()
    {
      task_types.insert(task_type.clone());
    }

    task_types.into_iter().collect()
  }
}

async fn run_handler_future(fut: BoxedHandlerFuture, task_id: String) -> HandlerResult {
  match tokio::spawn(fut).await {
    Ok(result) => result,
    Err(err) => HandlerResult::failure(format!("handler panic (task={task_id}): {err}")),
  }
}

fn normalize_options(opts: &mut WorkerOptions) {
  if opts.concurrency == 0 {
    opts.concurrency = DEFAULT_CONCURRENCY;
  }
  if opts.balance_channel_capacity == 0 {
    opts.balance_channel_capacity = DEFAULT_BALANCE_CHANNEL_CAPACITY;
  }
  if opts.heartbeat_interval.is_zero() {
    opts.heartbeat_interval = DEFAULT_HEARTBEAT_INTERVAL;
  }
  if opts.report_retry.max_attempts == 0 {
    opts.report_retry.max_attempts = DEFAULT_REPORT_RETRY_MAX_ATTEMPTS;
  }
  if opts.report_retry.initial_backoff.is_zero() {
    opts.report_retry.initial_backoff = DEFAULT_REPORT_RETRY_INITIAL_BACKOFF;
  }
  if opts.report_retry.max_backoff.is_zero() {
    opts.report_retry.max_backoff = DEFAULT_REPORT_RETRY_MAX_BACKOFF;
  }
  if opts.report_retry.multiplier <= 1.0 {
    opts.report_retry.multiplier = DEFAULT_REPORT_RETRY_MULTIPLIER;
  }
}

fn next_backoff(current: Duration, opts: &RetryOptions) -> Duration {
  let next = Duration::from_secs_f64(current.as_secs_f64() * opts.multiplier);
  if next > opts.max_backoff {
    opts.max_backoff
  } else {
    next
  }
}

fn task_from_proto(task: taskv1::Task) -> Task {
  Task {
    id: task.id,
    namespace: task.namespace,
    task_type: task.task_type,
    input_args: task.input_args.to_vec(),
    callback_url: task.callback_url,
    max_retry_cnt: task.max_retry_cnt,
    expire_time: task.expire_time,
    ctime: task.ctime,
    trace_id: task.trace_id,
  }
}

fn is_timeout_status(err: &tonic::Status) -> bool {
  err.code() == Code::DeadlineExceeded
}
