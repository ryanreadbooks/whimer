use tokio::sync::watch;

#[derive(Clone, Debug)]
pub struct Subscriber {
  key: String,
  rx: watch::Receiver<Vec<String>>,
}

impl Subscriber {
  pub(crate) fn new(key: String, rx: watch::Receiver<Vec<String>>) -> Self {
    Self { key, rx }
  }

  pub fn key(&self) -> &str {
    &self.key
  }

  pub fn values(&self) -> Vec<String> {
    self.rx.borrow().clone()
  }

  pub fn subscribe(&self) -> watch::Receiver<Vec<String>> {
    self.rx.clone()
  }

  pub async fn changed(&mut self) -> std::result::Result<Vec<String>, watch::error::RecvError> {
    self.rx.changed().await?;
    Ok(self.rx.borrow().clone())
  }
}
