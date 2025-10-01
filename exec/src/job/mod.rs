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
}

impl TradeRequest {
    pub fn is_buy(&self) -> bool {
        self.side.eq_ignore_ascii_case("buy")
    }
}

impl Intent {
    pub fn parse_trade(&self) -> anyhow::Result<TradeRequest> {
        let trade: TradeRequest = serde_json::from_value(self.payload.clone())?;
        Ok(trade)
    }

    pub fn age(&self) -> anyhow::Result<chrono::Duration> {
        let now = chrono::Utc::now();
        Ok(now - self.created_at)
    }
}
