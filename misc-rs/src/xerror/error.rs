use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Error {
  #[serde(rename = "stcode")]
  pub status_code: i32, // http响应状态码

  #[serde(rename = "code")]
  pub code: i32, // 业务响应码

  #[serde(rename = "msg")]
  pub message: String,
}

impl Default for Error {
  fn default() -> Self {
    Self {
      status_code: 500,
      code: 19999,
      message: "服务错误".to_string(),
    }
  }
}

impl Error {
  pub fn new(status_code: i32, code: i32, message: String) -> Self {
    Self {
      status_code,
      code,
      message,
    }
  }

  pub fn status_code(&self) -> i32 {
    self.status_code
  }

  pub fn code(&self) -> i32 {
    self.code
  }

  pub fn from_json(json: impl AsRef<str>) -> Self {
    serde_json::from_str(json.as_ref()).unwrap_or_default()
  }
}
