use super::subscriber::Subscriber;
use etcd_client::{Client, EventType, GetOptions, WatchOptions};
use std::collections::{HashMap, HashSet};
use std::fmt::{Display, Formatter};
use std::sync::{Arc, OnceLock};
use tokio::sync::{Mutex, RwLock, watch};
use tokio::time::{Duration, sleep};

const ENDPOINTS_SEPARATOR: &str = ",";
const KEY_DELIMITER: char = '/';
const RETRY_INTERVAL: Duration = Duration::from_secs(1);

pub type Result<T> = std::result::Result<T, DiscoveryError>;

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct KV {
  pub key: String,
  pub val: String,
}

#[derive(Debug)]
pub enum DiscoveryError {
  EmptyEtcdHosts,
  EmptyEtcdKey,
  Etcd(etcd_client::Error),
  WatchCanceled { key: String, reason: String },
  WatchStreamClosed { key: String },
}

impl Display for DiscoveryError {
  fn fmt(&self, f: &mut Formatter<'_>) -> std::fmt::Result {
    match self {
      Self::EmptyEtcdHosts => write!(f, "empty etcd hosts"),
      Self::EmptyEtcdKey => write!(f, "empty etcd key"),
      Self::Etcd(err) => write!(f, "etcd error: {err}"),
      Self::WatchCanceled { key, reason } => {
        write!(f, "watch canceled for key {key}: {reason}")
      }
      Self::WatchStreamClosed { key } => write!(f, "watch stream closed for key {key}"),
    }
  }
}

impl std::error::Error for DiscoveryError {}

impl From<etcd_client::Error> for DiscoveryError {
  fn from(err: etcd_client::Error) -> Self {
    Self::Etcd(err)
  }
}

#[derive(Clone, Debug, Default)]
pub struct Registry {
  inner: Arc<RegistryInner>,
}

#[derive(Debug, Default)]
struct RegistryInner {
  clusters: Mutex<HashMap<String, Arc<Cluster>>>,
}

#[derive(Debug)]
pub struct Cluster {
  endpoints: Vec<String>,
  key: String, // derived from endpoints
  state: RwLock<ClusterState>,
}

#[derive(Debug, Default)]
struct ClusterState {
  values: HashMap<String, HashMap<String, String>>, // watch_key -> etcd_key -> addr
  senders: HashMap<String, watch::Sender<Vec<String>>>, // watch_key -> subscribers
  watching: HashSet<String>,
}

pub fn global_registry() -> &'static Registry {
  static GLOBAL: OnceLock<Registry> = OnceLock::new();
  GLOBAL.get_or_init(Registry::default)
}

impl Registry {
  pub async fn monitor(
    &self,
    endpoints: Vec<String>,
    key: impl Into<String>,
  ) -> Result<Subscriber> {
    if endpoints.is_empty() {
      return Err(DiscoveryError::EmptyEtcdHosts);
    }

    let key = key.into();
    if key.is_empty() {
      return Err(DiscoveryError::EmptyEtcdKey);
    }

    let cluster = self.get_or_create_cluster(endpoints).await;
    cluster.monitor(key).await
  }

  async fn get_or_create_cluster(&self, endpoints: Vec<String>) -> Arc<Cluster> {
    let cluster_key = get_cluster_key(&endpoints);
    let mut clusters = self.inner.clusters.lock().await;

    if let Some(cluster) = clusters.get(&cluster_key) {
      return Arc::clone(cluster);
    }

    let cluster = Arc::new(Cluster::new(endpoints));
    clusters.insert(cluster_key, Arc::clone(&cluster));
    cluster
  }
}

impl Cluster {
  fn new(endpoints: Vec<String>) -> Self {
    let key = get_cluster_key(&endpoints);
    Self {
      endpoints,
      key,
      state: RwLock::new(ClusterState::default()),
    }
  }

  pub fn key(&self) -> &str {
    &self.key
  }

  async fn monitor(self: &Arc<Self>, key: String) -> Result<Subscriber> {
    let (sender, should_spawn) = {
      let mut state = self.state.write().await;

      let initial_values = state
        .values
        .get(&key)
        .map(unique_sorted_values)
        .unwrap_or_default();

      let sender = if let Some(existing) = state.senders.get(&key) {
        existing.clone()
      } else {
        let (tx, _rx) = watch::channel(initial_values);
        state.senders.insert(key.clone(), tx.clone());
        tx
      };

      let should_spawn = state.watching.insert(key.clone());
      (sender, should_spawn)
    };

    if should_spawn {
      let cluster = Arc::clone(self);
      let watch_key = key.clone();
      tokio::spawn(async move {
        cluster.watch_key_loop(watch_key).await;
      });
    }

    Ok(Subscriber::new(key, sender.subscribe()))
  }

  async fn watch_key_loop(self: Arc<Self>, key: String) {
    loop {
      if let Err(err) = self.run_watch_cycle(&key).await {
        eprintln!(
          "discovery watch cycle failed for cluster={}, key={key}: {err}",
          self.key
        );
      }
      sleep(RETRY_INTERVAL).await;
    }
  }

  async fn run_watch_cycle(&self, key: &str) -> Result<()> {
    let mut client = Client::connect(self.endpoints.clone(), None).await?;
    let mut revision = self.load_snapshot(&mut client, key).await?;
    let prefix = make_key_prefix(key);

    'recreate_watch: loop {
      let mut watch_opts = WatchOptions::new().with_prefix();
      if revision > 0 {
        watch_opts = watch_opts.with_start_revision(revision + 1);
      }

      let mut watch_stream = client.watch(prefix.as_bytes(), Some(watch_opts)).await?;
      while let Some(resp) = watch_stream.message().await? {
        if resp.canceled() {
          if resp.compact_revision() > 0 {
            revision = self.load_snapshot(&mut client, key).await?;
            continue 'recreate_watch;
          }

          return Err(DiscoveryError::WatchCanceled {
            key: key.to_string(),
            reason: resp.cancel_reason().to_string(),
          });
        }

        self.apply_watch_events(key, resp.events()).await;
      }

      return Err(DiscoveryError::WatchStreamClosed {
        key: key.to_string(),
      });
    }
  }

  async fn load_snapshot(&self, client: &mut Client, key: &str) -> Result<i64> {
    let prefix = make_key_prefix(key);
    let resp = client
      .get(prefix.as_bytes(), Some(GetOptions::new().with_prefix()))
      .await?;

    let mut values = HashMap::new();
    for kv in resp.kvs() {
      values.insert(
        bytes_to_string_lossy(kv.key()),
        bytes_to_string_lossy(kv.value()),
      );
    }

    let revision = resp.header().map(|h| h.revision()).unwrap_or_default();
    self.publish_values(key, values).await;

    Ok(revision)
  }

  async fn apply_watch_events(&self, key: &str, events: &[etcd_client::Event]) {
    let (sender, snapshot, changed) = {
      let mut state = self.state.write().await;
      let mut changed = false;
      let snapshot = {
        let values = state.values.entry(key.to_string()).or_default();

        for event in events {
          let Some(kv) = event.kv() else {
            continue;
          };

          let etcd_key = bytes_to_string_lossy(kv.key());
          match event.event_type() {
            EventType::Put => {
              let addr = bytes_to_string_lossy(kv.value());
              let previous = values.insert(etcd_key, addr.clone());
              if previous.as_deref() != Some(addr.as_str()) {
                changed = true;
              }
            }
            EventType::Delete => {
              if values.remove(&etcd_key).is_some() {
                changed = true;
              }
            }
          }
        }

        unique_sorted_values(values)
      };

      let sender = state.senders.get(key).cloned();
      (sender, snapshot, changed)
    };

    if changed {
      if let Some(tx) = sender {
        let _ = tx.send(snapshot);
      }
    }
  }

  async fn publish_values(&self, key: &str, values: HashMap<String, String>) {
    let (sender, snapshot) = {
      let mut state = self.state.write().await;
      state.values.insert(key.to_string(), values);

      let snapshot = state
        .values
        .get(key)
        .map(unique_sorted_values)
        .unwrap_or_default();

      let sender = state.senders.get(key).cloned();
      (sender, snapshot)
    };

    if let Some(tx) = sender {
      let _ = tx.send(snapshot);
    }
  }
}

fn get_cluster_key(endpoints: &[String]) -> String {
  let mut endpoints = endpoints.to_vec();
  endpoints.sort();
  endpoints.join(ENDPOINTS_SEPARATOR)
}

fn make_key_prefix(key: &str) -> String {
  if key.ends_with(KEY_DELIMITER) {
    key.to_string()
  } else {
    format!("{key}{KEY_DELIMITER}")
  }
}

fn bytes_to_string_lossy(bytes: &[u8]) -> String {
  String::from_utf8_lossy(bytes).into_owned()
}

fn unique_sorted_values(values: &HashMap<String, String>) -> Vec<String> {
  let mut addresses: Vec<String> = values
    .values()
    .cloned()
    .collect::<HashSet<_>>()
    .into_iter()
    .collect();
  addresses.sort();
  addresses
}
