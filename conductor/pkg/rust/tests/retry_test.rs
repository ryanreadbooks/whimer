use conductor_sdk_rust::{producer, worker};
use serde::{Deserialize, Serialize};
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::time::{Duration, SystemTime};
use tokio::sync::mpsc;
use tokio::time::{sleep, timeout};

mod common;
use common::{
  ensure_local_etcd_ready, new_producer, new_worker, spawn_worker, stop_worker, unique_task_type,
  wait_task_state,
};

#[derive(Debug, Clone, Serialize, Deserialize)]
struct RetryOutput {
  attempt: usize,
}

#[tokio::test]
async fn test_retry_and_success() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");
  let worker_client = new_worker("retry-success", None, 1)
    .await
    .expect("create worker should succeed");

  let task_type = unique_task_type("rust_retry_success");
  let execution_count = Arc::new(AtomicUsize::new(0));
  let execution_count_ref = Arc::clone(&execution_count);
  let (done_tx, mut done_rx) = mpsc::unbounded_channel::<String>();

  worker_client.register_handler(task_type.clone(), move |task| {
    let done_tx = done_tx.clone();
    let execution_count = Arc::clone(&execution_count);
    async move {
      let attempt = execution_count.fetch_add(1, Ordering::SeqCst) + 1;
      if attempt == 1 {
        return worker::HandlerResult::retryable_failure("simulated failure at first attempt");
      }

      let _ = done_tx.send(task.id.clone());
      worker::HandlerResult::success_json(&RetryOutput { attempt })
        .expect("serialize retry output should succeed")
    }
  });

  let mut worker_handle = spawn_worker(worker_client.clone());
  sleep(Duration::from_secs(1)).await;

  let task_id = producer_client
    .schedule(
      task_type,
      Some(&serde_json::json!({"case":"retry_success"})),
      producer::ScheduleOptions {
        max_retry: 3,
        expire_after: Some(Duration::from_secs(120)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("schedule task should succeed");

  let done_task_id = timeout(Duration::from_secs(40), done_rx.recv())
    .await
    .expect("wait retry success timeout")
    .expect("done channel closed unexpectedly");
  assert_eq!(done_task_id, task_id);

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["success", "failure", "expired"],
    Duration::from_secs(30),
  )
  .await;
  assert_eq!(final_task.state, "success");

  let output = final_task
    .unmarshal_output::<RetryOutput>()
    .expect("unmarshal retry output should succeed")
    .expect("retry output should exist");
  assert!(output.attempt >= 2);
  assert!(execution_count_ref.load(Ordering::SeqCst) >= 2);

  stop_worker(&worker_client, &mut worker_handle).await;
}

#[tokio::test]
async fn test_failure_without_retry() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");
  let worker_client = new_worker("failure-no-retry", None, 1)
    .await
    .expect("create worker should succeed");

  let task_type = unique_task_type("rust_failure_no_retry");
  let (seen_tx, mut seen_rx) = mpsc::unbounded_channel::<String>();

  worker_client.register_handler(task_type.clone(), move |task| {
    let seen_tx = seen_tx.clone();
    async move {
      let _ = seen_tx.send(task.id.clone());
      worker::HandlerResult::failure("simulated failure without retry")
    }
  });

  let mut worker_handle = spawn_worker(worker_client.clone());
  sleep(Duration::from_secs(1)).await;

  let task_id = producer_client
    .schedule(
      task_type,
      Some(&serde_json::json!({"case":"failure_without_retry"})),
      producer::ScheduleOptions {
        max_retry: 0,
        expire_after: Some(Duration::from_secs(60)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("schedule task should succeed");

  let seen_task_id = timeout(Duration::from_secs(15), seen_rx.recv())
    .await
    .expect("wait first execution timeout")
    .expect("seen channel closed unexpectedly");
  assert_eq!(seen_task_id, task_id);

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["failure", "expired"],
    Duration::from_secs(20),
  )
  .await;
  assert_eq!(final_task.state, "failure");

  stop_worker(&worker_client, &mut worker_handle).await;
}

#[tokio::test]
async fn test_max_retry_exhausted() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");
  let worker_client = new_worker("max-retry-exhausted", None, 1)
    .await
    .expect("create worker should succeed");

  let task_type = unique_task_type("rust_max_retry_exhausted");
  let execution_count = Arc::new(AtomicUsize::new(0));
  let execution_count_ref = Arc::clone(&execution_count);

  worker_client.register_handler(task_type.clone(), move |_task| {
    let execution_count = Arc::clone(&execution_count);
    async move {
      execution_count.fetch_add(1, Ordering::SeqCst);
      worker::HandlerResult::retryable_failure("simulated failure until retry exhausted")
    }
  });

  let mut worker_handle = spawn_worker(worker_client.clone());
  sleep(Duration::from_secs(1)).await;

  let max_retry = 2i64;
  let expected_executions = (max_retry + 1) as usize;
  let task_id = producer_client
    .schedule(
      task_type,
      Some(&serde_json::json!({"case":"max_retry_exhausted"})),
      producer::ScheduleOptions {
        max_retry,
        expire_after: Some(Duration::from_secs(120)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("schedule task should succeed");

  let start = SystemTime::now();
  loop {
    if execution_count_ref.load(Ordering::SeqCst) >= expected_executions {
      break;
    }
    assert!(
      start
        .elapsed()
        .expect("elapsed time should be available")
        <= Duration::from_secs(120),
      "wait retry exhaustion execution timeout, got={}",
      execution_count_ref.load(Ordering::SeqCst)
    );
    sleep(Duration::from_secs(1)).await;
  }

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["failure", "expired"],
    Duration::from_secs(30),
  )
  .await;
  assert!(
    final_task.state == "failure" || final_task.state == "expired",
    "unexpected final state: {}",
    final_task.state
  );
  assert!(execution_count_ref.load(Ordering::SeqCst) >= expected_executions);

  stop_worker(&worker_client, &mut worker_handle).await;
}

#[tokio::test]
async fn test_retry_and_timeout() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");
  let worker_client = new_worker("retry-timeout", None, 1)
    .await
    .expect("create worker should succeed");

  let task_type = unique_task_type("rust_retry_timeout");
  let execution_count = Arc::new(AtomicUsize::new(0));
  let execution_count_ref = Arc::clone(&execution_count);
  worker_client.register_handler(task_type.clone(), move |_task| {
    let execution_count = Arc::clone(&execution_count);
    async move {
      execution_count.fetch_add(1, Ordering::SeqCst);
      worker::HandlerResult::retryable_failure("always fail until timeout")
    }
  });

  let mut worker_handle = spawn_worker(worker_client.clone());
  sleep(Duration::from_secs(1)).await;

  let task_id = producer_client
    .schedule(
      task_type,
      Some(&serde_json::json!({"case":"retry_timeout"})),
      producer::ScheduleOptions {
        max_retry: -1,
        expire_after: Some(Duration::from_secs(10)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("schedule task should succeed");

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["expired", "failure"],
    Duration::from_secs(35),
  )
  .await;
  assert!(
    final_task.state == "expired" || final_task.state == "failure",
    "unexpected final state: {}",
    final_task.state
  );
  assert!(
    execution_count_ref.load(Ordering::SeqCst) >= 2,
    "expected at least 2 executions, got {}",
    execution_count_ref.load(Ordering::SeqCst)
  );

  stop_worker(&worker_client, &mut worker_handle).await;
}

#[tokio::test]
async fn test_no_retry_and_timeout() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");

  let task_type = unique_task_type("rust_no_retry_timeout");
  let task_id = producer_client
    .schedule(
      task_type,
      Some(&serde_json::json!({"case":"no_retry_timeout"})),
      producer::ScheduleOptions {
        max_retry: 0,
        expire_after: Some(Duration::from_secs(10)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("schedule task should succeed");

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["expired", "failure"],
    Duration::from_secs(35),
  )
  .await;
  assert!(
    final_task.state == "expired" || final_task.state == "failure",
    "unexpected final state: {}",
    final_task.state
  );
}
