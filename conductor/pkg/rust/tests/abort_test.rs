use conductor_sdk_rust::{producer, worker};
use std::sync::atomic::{AtomicI64, Ordering};
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::mpsc;
use tokio::time::{sleep, timeout};

mod common;
use common::{
  ensure_local_etcd_ready, new_producer, new_worker, spawn_worker, stop_worker, unique_task_type,
  wait_task_state,
};

#[tokio::test]
async fn test_abort_with_task_handler() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");
  let worker_client = new_worker("abort-task-handler", Some(Duration::from_secs(2)), 1)
    .await
    .expect("create worker should succeed");

  let task_type = unique_task_type("rust_abort_task_handler");
  let (started_tx, mut started_rx) = mpsc::unbounded_channel::<String>();
  let (aborted_tx, mut aborted_rx) = mpsc::unbounded_channel::<String>();

  worker_client.register_task_handler(task_type.clone(), move |tc| {
    let started_tx = started_tx.clone();
    let aborted_tx = aborted_tx.clone();
    async move {
      let task_id = tc.task().id.clone();
      let _ = started_tx.send(task_id.clone());
      let _ = timeout(Duration::from_secs(40), tc.wait_abort()).await;
      if tc.is_aborted() {
        let _ = aborted_tx.send(task_id);
      }
      worker::HandlerResult::failure("task aborted by conductor")
    }
  });

  let mut worker_handle = spawn_worker(worker_client.clone());
  sleep(Duration::from_secs(1)).await;

  let task_id = producer_client
    .schedule(
      task_type,
      Some(&serde_json::json!({"case":"abort_with_task_handler"})),
      producer::ScheduleOptions {
        max_retry: 0,
        expire_after: Some(Duration::from_secs(60)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("schedule task should succeed");

  let started_task_id = timeout(Duration::from_secs(15), started_rx.recv())
    .await
    .expect("wait task start timeout")
    .expect("start channel closed unexpectedly");
  assert_eq!(started_task_id, task_id);

  sleep(Duration::from_secs(3)).await;
  producer_client
    .abort_task(task_id.clone())
    .await
    .expect("abort task should succeed");

  let aborted_task_id = timeout(Duration::from_secs(15), aborted_rx.recv())
    .await
    .expect("wait abort signal timeout")
    .expect("abort channel closed unexpectedly");
  assert_eq!(aborted_task_id, task_id);

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["failure", "aborted"],
    Duration::from_secs(20),
  )
  .await;
  assert!(
    final_task.state == "failure" || final_task.state == "aborted",
    "unexpected final state: {}",
    final_task.state
  );

  stop_worker(&worker_client, &mut worker_handle).await;
}

#[tokio::test]
async fn test_abort_with_progress_provider() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");
  let worker_client = new_worker("abort-progress-provider", Some(Duration::from_secs(1)), 1)
    .await
    .expect("create worker should succeed");

  let task_type = unique_task_type("rust_abort_progress_provider");
  let (started_tx, mut started_rx) = mpsc::unbounded_channel::<String>();
  let (aborted_tx, mut aborted_rx) = mpsc::unbounded_channel::<String>();
  let progress = Arc::new(AtomicI64::new(0));
  let progress_for_assert = Arc::clone(&progress);

  worker_client.register_task_handler(task_type.clone(), move |tc| {
    let started_tx = started_tx.clone();
    let aborted_tx = aborted_tx.clone();
    let progress = Arc::clone(&progress);
    async move {
      let task_id = tc.task().id.clone();
      let _ = started_tx.send(task_id.clone());

      let progress_for_provider = Arc::clone(&progress);
      tc.set_progress_provider(move || progress_for_provider.load(Ordering::SeqCst));

      for i in 0..=200 {
        progress.store(i, Ordering::SeqCst);
        if tc.is_aborted() {
          let _ = aborted_tx.send(task_id);
          return worker::HandlerResult::failure("task aborted by conductor");
        }
        sleep(Duration::from_millis(100)).await;
      }

      worker::HandlerResult::success()
    }
  });

  let mut worker_handle = spawn_worker(worker_client.clone());
  sleep(Duration::from_secs(1)).await;

  let task_id = producer_client
    .schedule(
      task_type,
      Some(&serde_json::json!({"case":"abort_with_progress_provider"})),
      producer::ScheduleOptions {
        max_retry: 0,
        expire_after: Some(Duration::from_secs(60)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("schedule task should succeed");

  timeout(Duration::from_secs(15), started_rx.recv())
    .await
    .expect("wait task start timeout")
    .expect("start channel closed unexpectedly");

  sleep(Duration::from_secs(3)).await;
  producer_client
    .abort_task(task_id.clone())
    .await
    .expect("abort task should succeed");

  timeout(Duration::from_secs(15), aborted_rx.recv())
    .await
    .expect("wait abort signal timeout")
    .expect("abort channel closed unexpectedly");

  let final_progress = progress_for_assert.load(Ordering::SeqCst);
  assert!(
    (1..200).contains(&final_progress),
    "unexpected final progress: {}",
    final_progress
  );

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["failure", "aborted"],
    Duration::from_secs(20),
  )
  .await;
  assert!(
    final_task.state == "failure" || final_task.state == "aborted",
    "unexpected final state: {}",
    final_task.state
  );

  stop_worker(&worker_client, &mut worker_handle).await;
}

#[tokio::test]
async fn test_manual_report_progress_abort() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");
  let worker_client = new_worker("manual-report-progress", Some(Duration::from_secs(10)), 1)
    .await
    .expect("create worker should succeed");

  let task_type = unique_task_type("rust_manual_report_progress");
  let (started_tx, mut started_rx) = mpsc::unbounded_channel::<String>();
  let (aborted_tx, mut aborted_rx) = mpsc::unbounded_channel::<String>();

  worker_client.register_task_handler(task_type.clone(), move |tc| {
    let started_tx = started_tx.clone();
    let aborted_tx = aborted_tx.clone();
    async move {
      let task_id = tc.task().id.clone();
      let _ = started_tx.send(task_id.clone());

      for progress in 0..=200 {
        if tc.report_progress(progress).await {
          let _ = aborted_tx.send(task_id);
          return worker::HandlerResult::failure("task aborted by conductor");
        }
        sleep(Duration::from_millis(100)).await;
      }

      worker::HandlerResult::success()
    }
  });

  let mut worker_handle = spawn_worker(worker_client.clone());
  sleep(Duration::from_secs(1)).await;

  let task_id = producer_client
    .schedule(
      task_type,
      Some(&serde_json::json!({"case":"manual_report_progress_abort"})),
      producer::ScheduleOptions {
        max_retry: 0,
        expire_after: Some(Duration::from_secs(60)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("schedule task should succeed");

  timeout(Duration::from_secs(15), started_rx.recv())
    .await
    .expect("wait task start timeout")
    .expect("start channel closed unexpectedly");

  sleep(Duration::from_secs(2)).await;
  producer_client
    .abort_task(task_id.clone())
    .await
    .expect("abort task should succeed");

  timeout(Duration::from_secs(10), aborted_rx.recv())
    .await
    .expect("wait abort signal timeout")
    .expect("abort channel closed unexpectedly");

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["failure", "aborted"],
    Duration::from_secs(20),
  )
  .await;
  assert!(
    final_task.state == "failure" || final_task.state == "aborted",
    "unexpected final state: {}",
    final_task.state
  );

  stop_worker(&worker_client, &mut worker_handle).await;
}
