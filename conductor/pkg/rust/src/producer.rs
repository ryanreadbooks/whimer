use whimer_idl_rust::conductor::api::taskservice::v1::task_service_client::TaskServiceClient;

#[derive(Debug, Default)]
pub struct ClientOptions {
  pub hosts: Vec<String>,
  pub host_key: String,
  pub namespace: String,
}

impl ClientOptions {
  pub fn new(hosts: Vec<String>, host_key: &'static str, namespace: &'static str) -> Self {
    Self {
      hosts,
      host_key: host_key.to_string(),
      namespace: namespace.to_string(),
    }
  }
}

#[derive(Debug)]
pub struct Client {
  pub opts: ClientOptions,
}
