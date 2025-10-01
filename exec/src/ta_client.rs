use anyhow::Result;
use reqwest::Client;
use serde::Deserialize;
use std::collections::HashMap;

#[derive(Clone)]
pub struct TAClient {
    base_url: String,
    client: Client,
}

impl TAClient {
    pub fn new(base_url: String) -> Self {
        Self {
            base_url: base_url.trim_end_matches('/').to_string(),
            client: Client::new(),
        }
    }

    pub async fn rsi(&self, pair: &str, interval: &str) -> Result<f64> {
        let url = format!("{}/v1/indicators/rsi/{}/{}", self.base_url, pair, interval);
        let resp: IndicatorValue = self
            .client
            .get(url)
            .send()
            .await?
            .error_for_status()?
            .json()
            .await?;
        Ok(resp.value)
    }

    pub async fn signals(&self, pair: &str, interval: &str) -> Result<HashMap<String, f64>> {
        let url = format!(
            "{}/v1/indicators/signals/{}/{}",
            self.base_url, pair, interval
        );
        let resp: SignalResponse = self
            .client
            .get(url)
            .send()
            .await?
            .error_for_status()?
            .json()
            .await?;
        Ok(resp.signals)
    }
}

#[derive(Deserialize)]
struct IndicatorValue {
    value: f64,
}

#[derive(Deserialize)]
struct SignalResponse {
    signals: HashMap<String, f64>,
}
