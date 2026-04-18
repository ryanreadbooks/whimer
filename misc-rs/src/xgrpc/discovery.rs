use crate::discov::{DiscoveryError as RegistryError, Subscriber, global_registry};
use std::collections::HashSet;
use std::fmt::{Display, Formatter};
use tokio::sync::mpsc::Sender;
use tokio::time::{Duration, sleep};
use tonic::transport::{Channel, Endpoint, channel::Change};

const RESUBSCRIBE_RETRY_INTERVAL: Duration = Duration::from_secs(1);

#[derive(Debug, Default, Clone)]
pub struct EtcdConfig {
  hosts: Vec<String>, // etcd host
  key: String,        // target key
}

impl EtcdConfig {
  pub fn new(hosts: Vec<String>, key: impl Into<String>) -> Self {
    Self {
      hosts,
      key: key.into(),
    }
  }

  pub fn hosts(&self) -> &[String] {
    &self.hosts
  }

  pub fn key(&self) -> &str {
    &self.key
  }
}

#[derive(Debug)]
pub enum DiscoveryError {
  Registry(RegistryError),
  BalanceChannelClosed,
}

impl Display for DiscoveryError {
  fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
    match self {
      Self::Registry(err) => write!(f, "{err}"),
      Self::BalanceChannelClosed => write!(f, "tonic balance channel closed"),
    }
  }
}

impl std::error::Error for DiscoveryError {}

impl From<RegistryError> for DiscoveryError {
  fn from(value: RegistryError) -> Self {
    Self::Registry(value)
  }
}

pub type Result<T> = std::result::Result<T, DiscoveryError>;

pub struct EtcdDiscovery {
  config: EtcdConfig,
  subscriber: Subscriber,
}

impl EtcdDiscovery {
  pub async fn new(config: EtcdConfig) -> Result<Self> {
    let subscriber = global_registry()
      .monitor(config.hosts.clone(), config.key.clone())
      .await?;

    Ok(Self { config, subscriber })
  }

  pub fn config(&self) -> &EtcdConfig {
    &self.config
  }

  pub fn instances(&self) -> Vec<String> {
    self.subscriber.values()
  }

  pub fn subscribe(&self) -> tokio::sync::watch::Receiver<Vec<String>> {
    self.subscriber.subscribe()
  }

  pub async fn balanced_channel(&self, capacity: usize) -> Result<Channel> {
    let (channel, tx) = Channel::balance_channel::<String>(capacity);
    let mut known = HashSet::new();

    sync_instances(&tx, &mut known, self.instances()).await?;

    let mut rx = self.subscribe();
    let config = self.config.clone();
    tokio::spawn(async move {
      loop {
        match rx.changed().await {
          Ok(()) => {
            let next = rx.borrow().clone();
            if let Err(err) = sync_instances(&tx, &mut known, next).await {
              eprintln!("xgrpc discovery sync failed: {err}");
              if matches!(err, DiscoveryError::BalanceChannelClosed) {
                break;
              }
            }
          }
          Err(_) => {
            eprintln!("xgrpc discovery receiver closed, trying to resubscribe");
            match global_registry()
              .monitor(config.hosts.clone(), config.key.clone())
              .await
            {
              Ok(subscriber) => {
                if let Err(err) = sync_instances(&tx, &mut known, subscriber.values()).await {
                  eprintln!("xgrpc discovery resubscribe sync failed: {err}");
                  if matches!(err, DiscoveryError::BalanceChannelClosed) {
                    break;
                  }
                }
                rx = subscriber.subscribe();
              }
              Err(err) => {
                eprintln!("xgrpc discovery resubscribe failed: {err}");
                sleep(RESUBSCRIBE_RETRY_INTERVAL).await;
              }
            }
          }
        }
      }
    });

    Ok(channel)
  }
}

async fn sync_instances(
  tx: &Sender<Change<String, Endpoint>>,
  current: &mut HashSet<String>,
  instances: Vec<String>,
) -> Result<()> {
  let next = instances
    .iter()
    .map(|endpoint| normalize_endpoint(endpoint))
    .collect::<HashSet<_>>();

  let mut removed = current.difference(&next).cloned().collect::<Vec<_>>();
  removed.sort();

  for endpoint in removed {
    tx.send(Change::Remove(endpoint.clone()))
      .await
      .map_err(|_| DiscoveryError::BalanceChannelClosed)?;
    current.remove(&endpoint);
  }

  let mut added = next.difference(current).cloned().collect::<Vec<_>>();
  added.sort();

  for endpoint in added {
    let channel_endpoint = match Endpoint::from_shared(endpoint.clone()) {
      Ok(channel_endpoint) => channel_endpoint,
      Err(_) => {
        eprintln!("xgrpc discovery ignore invalid endpoint: {endpoint}");
        continue;
      }
    };

    tx.send(Change::Insert(endpoint.clone(), channel_endpoint))
      .await
      .map_err(|_| DiscoveryError::BalanceChannelClosed)?;

    current.insert(endpoint);
  }

  Ok(())
}

fn normalize_endpoint(endpoint: &str) -> String {
  if endpoint.starts_with("http://") || endpoint.starts_with("https://") {
    endpoint.to_string()
  } else {
    format!("http://{endpoint}")
  }
}
