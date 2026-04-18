#![allow(dead_code)]

use conductor_sdk_rust::{producer, worker};
use std::sync::atomic::{AtomicU64, Ordering};
use std::time::{Duration, SystemTime, UNIX_EPOCH};
use tokio::net::TcpStream;
use tokio::task::JoinHandle;
use tokio::time::{sleep, timeout};

const TEST_HOSTS: [&str; 1] = ["localhost:2379"];
const TEST_HOST_KEY: &str = "whimer.conductor.rpc";
const TEST_NAMESPACE: &str = "default";
static UNIQUE_COUNTER: AtomicU64 = AtomicU64::new(0);

pub fn integration_hosts() -> Vec<String> {
  TEST_HOSTS.iter().map(|item| (*item).to_string()).collect()
}

pub fn test_host_key() -> &'static str {
  TEST_HOST_KEY
}

pub fn test_namespace() -> &'static str {
  TEST_NAMESPACE
}

pub fn unique_suffix() -> String {
  let now = SystemTime::now()
    .duration_since(UNIX_EPOCH)
    .expect("system time should be later than unix epoch")
    .as_millis();
  let seq = UNIQUE_COUNTER.fetch_add(1, Ordering::SeqCst);
  format!("{now}_{seq}")
}

pub fn unique_task_type(prefix: &str) -> String {
  format!("{prefix}_{}", unique_suffix())
}

pub async fn ensure_local_etcd_ready() {
  let etcd_addr = "127.0.0.1:2379";
  match timeout(Duration::from_secs(1), TcpStream::connect(etcd_addr)).await {
    Ok(Ok(_)) => {}
    Ok(Err(err)) => panic!(
      "integration test requires etcd at {}, connection failed: {}",
      etcd_addr, err
    ),
    Err(_) => panic!(
      "integration test requires etcd at {}, connection timeout",
      etcd_addr
    ),
  }
}

pub async fn new_producer() -> producer::Result<producer::Client> {
  producer::Client::new(producer::ClientOptions::new(
    integration_hosts(),
    TEST_HOST_KEY,
    TEST_NAMESPACE,
  ))
  .await
}

pub async fn new_worker(
  worker_id_prefix: &str,
  heartbeat_interval: Option<Duration>,
  concurrency: usize,
) -> worker::Result<worker::Worker> {
  let mut worker_opts = worker::WorkerOptions::new(integration_hosts(), TEST_HOST_KEY);
  worker_opts.worker_id = format!("{worker_id_prefix}-{}", unique_suffix());
  worker_opts.ip = "127.0.0.1".to_string();
  worker_opts.concurrency = concurrency;
  if let Some(interval) = heartbeat_interval {
    worker_opts.heartbeat_interval = interval;
  }

  worker::Worker::new(worker_opts).await
}

pub async fn wait_task_state(
  producer_client: &producer::Client,
  task_id: &str,
  expected_states: &[&str],
  wait_timeout: Duration,
) -> producer::Task {
  let start = SystemTime::now();
  loop {
    let task = producer_client
      .get_task(task_id.to_string())
      .await
      .expect("get task should succeed")
      .expect("task should exist");

    if expected_states.iter().any(|state| task.state == *state) {
      return task;
    }

    let elapsed = start.elapsed().expect("elapsed time should be available");
    assert!(
      elapsed <= wait_timeout,
      "wait task state timeout, task_id={}, current_state={}, expected={:?}",
      task_id,
      task.state,
      expected_states
    );

    sleep(Duration::from_millis(300)).await;
  }
}

pub fn spawn_worker(worker_client: worker::Worker) -> JoinHandle<()> {
  tokio::spawn(async move {
    let _ = worker_client.run().await;
  })
}

pub async fn stop_worker(worker_client: &worker::Worker, handle: &mut JoinHandle<()>) {
  worker_client.stop();
  if timeout(Duration::from_secs(5), &mut *handle).await.is_err() {
    handle.abort();
  }
}
