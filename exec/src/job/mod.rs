use serde::{Deserialize, Serialize};
use serde_json::Value;
use std::time::SystemTime;

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Intent {
    pub id: String,
    pub principal: String,
    pub payload: Value,
    pub created_at: chrono::DateTime<chrono::Utc>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct TradeRequest {
    pub mode: String,
    pub token: String,
    pub size: f64,
    pub slippage_bps: i64,
    pub side: String,
    pub trigger: String,
    pub paper_trading: bool,
    pub interval: Option<String>,
    pub force: Option<bool>,
}

impl TradeRequest {
    pub fn is_buy(&self) -> bool {
        self.side.eq_ignore_ascii_case("buy")
    }

    pub fn interval(&self) -> &str {
        self.interval.as_deref().unwrap_or("1m")
    }

    pub fn force(&self) -> bool {
        self.force.unwrap_or(false)
    }
}

impl Intent {
    pub fn parse_trade(&self) -> anyhow::Result<TradeRequest> {
        let trade: TradeRequest = serde_json::from_value(self.payload.clone())?;
        Ok(trade)
    }

    pub fn parse_action(&self) -> Option<(String, Value)> {
        if let Some(action) = self.payload.get("action").and_then(|v| v.as_str()) {
            let payload = self
                .payload
                .get("payload")
                .cloned()
                .unwrap_or_else(|| Value::Object(serde_json::Map::new()));
            return Some((action.to_string(), payload));
        }
        None
    }

    pub fn age(&self) -> anyhow::Result<chrono::Duration> {
        let now = chrono::Utc::now();
        Ok(now - self.created_at)
    }
}
