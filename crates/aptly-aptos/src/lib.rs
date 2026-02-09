use anyhow::{anyhow, Context, Result};
use reqwest::blocking::{Client, Response};
use reqwest::StatusCode;
use serde_json::Value;

pub struct AptosClient {
    base_url: String,
    http: Client,
}

impl AptosClient {
    pub fn new(base_url: &str) -> Result<Self> {
        let base_url = base_url.trim().trim_end_matches('/').to_owned();
        if base_url.is_empty() {
            return Err(anyhow!("rpc url cannot be empty"));
        }

        let http = Client::builder()
            .build()
            .context("failed to build HTTP client")?;
        Ok(Self { base_url, http })
    }

    pub fn get_json(&self, path: &str) -> Result<Value> {
        let url = self.endpoint(path);
        let response = self
            .http
            .get(&url)
            .send()
            .with_context(|| format!("request failed: GET {url}"))?;
        self.handle_response(response)
    }

    pub fn post_json(&self, path: &str, body: &Value) -> Result<Value> {
        let url = self.endpoint(path);
        let response = self
            .http
            .post(&url)
            .json(body)
            .send()
            .with_context(|| format!("request failed: POST {url}"))?;
        self.handle_response(response)
    }

    fn endpoint(&self, path: &str) -> String {
        format!("{}/{}", self.base_url, path.trim_start_matches('/'))
    }

    fn handle_response(&self, response: Response) -> Result<Value> {
        let status = response.status();
        let text = response.text().context("failed to read response body")?;

        if status != StatusCode::OK && status != StatusCode::ACCEPTED {
            return Err(anyhow!("API error (status {}): {}", status.as_u16(), text));
        }

        serde_json::from_str(&text).context("failed to parse response JSON")
    }
}
