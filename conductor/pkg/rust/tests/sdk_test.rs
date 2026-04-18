use conductor_sdk_rust::{producer, worker};
use serde::{Deserialize, Serialize};
use std::time::Duration;
use tokio::sync::mpsc;
use tokio::time::{sleep, timeout};

mod common;
use common::{
  ensure_local_etcd_ready, new_producer, new_worker, spawn_worker, stop_worker, test_namespace,
  unique_task_type, wait_task_state,
};

#[derive(Debug, Clone, Serialize, Deserialize)]
struct EmailInput {
  to: String,
  subject: String,
  body: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct EmailOutput {
  message_id: String,
  status: String,
}

#[tokio::test]
async fn test_producer_register_task() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");

  let task_type = unique_task_type("rust_integration_register");
  let input = EmailInput {
    to: "test@example.com".to_string(),
    subject: "Hello From Rust".to_string(),
    body: "This is an integration test".to_string(),
  };

  let task_id = producer_client
    .schedule(
      task_type.clone(),
      Some(&input),
      producer::ScheduleOptions {
        namespace: Some(test_namespace().to_string()),
        max_retry: 1,
        expire_after: Some(Duration::from_secs(60)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("register task should succeed");

  let task = producer_client
    .get_task(task_id.clone())
    .await
    .expect("get task should succeed")
    .expect("task should exist");

  assert_eq!(task.id, task_id);
  assert_eq!(task.namespace, test_namespace());
  assert_eq!(task.task_type, task_type);
}

#[tokio::test]
async fn test_producer_and_worker_integration() {
  ensure_local_etcd_ready().await;
  let producer_client = new_producer()
    .await
    .expect("create producer client should succeed");

  let worker_client = new_worker("rust-worker", None, 1)
    .await
    .expect("create worker should succeed");

  let task_type = unique_task_type("rust_integration_worker");
  let (done_tx, mut done_rx) = mpsc::unbounded_channel::<String>();
  worker_client.register_handler(task_type.clone(), move |task| {
    let done_tx = done_tx.clone();
    async move {
      let input = match task.unmarshal_input::<EmailInput>() {
        Ok(Some(input)) => input,
        Ok(None) => return worker::HandlerResult::failure("missing input"),
        Err(err) => return worker::HandlerResult::failure(err.to_string()),
      };

      let output = EmailOutput {
        message_id: format!("msg_{}", task.id),
        status: format!("sent:{}", input.to),
      };
      let _ = done_tx.send(task.id.clone());

      match worker::HandlerResult::success_json(&output) {
        Ok(result) => result,
        Err(err) => worker::HandlerResult::failure(err.to_string()),
      }
    }
  });

  let mut worker_handle = spawn_worker(worker_client.clone());

  sleep(Duration::from_secs(1)).await;

  let input = EmailInput {
    to: "integration@example.com".to_string(),
    subject: "Worker Execute".to_string(),
    body: "Producer/Worker integration".to_string(),
  };
  let task_id = producer_client
    .schedule(
      task_type,
      Some(&input),
      producer::ScheduleOptions {
        max_retry: 1,
        expire_after: Some(Duration::from_secs(90)),
        ..producer::ScheduleOptions::default()
      },
    )
    .await
    .expect("register task should succeed");

  let completed_task_id = timeout(Duration::from_secs(20), done_rx.recv())
    .await
    .expect("wait worker execution timeout")
    .expect("completion channel should not close");
  assert_eq!(completed_task_id, task_id);

  let final_task = wait_task_state(
    &producer_client,
    &task_id,
    &["success", "failure", "expired", "aborted"],
    Duration::from_secs(20),
  )
  .await;
  assert_eq!(final_task.state, "success");

  let output = final_task
    .unmarshal_output::<EmailOutput>()
    .expect("unmarshal output should succeed")
    .expect("task output should exist");
  assert_eq!(output.message_id, format!("msg_{}", task_id));
  assert_eq!(output.status, "sent:integration@example.com");

  stop_worker(&worker_client, &mut worker_handle).await;
}
