use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

use anyhow::{anyhow, Result};

use crate::ta_client::TAClient;

#[derive(Debug, Clone)]
pub struct AutoTradeFilter {
    pub expression: String,
    pub interval: String,
}

impl AutoTradeFilter {
    pub async fn evaluate(&self, client: &TAClient, pair: &str) -> Result<bool> {
        let signals = client.signals(pair, &self.interval).await?;
        parse_expression(&self.expression, &signals)
    }
}

fn parse_expression(expr: &str, values: &HashMap<String, f64>) -> Result<bool> {
    let expr = expr.trim().to_lowercase();
    // Currently support format indicator<value or indicator>value
    if let Some(idx) = expr.find('<') {
        let (lhs, rhs) = expr.split_at(idx);
        let indicator = lhs.trim();
        let threshold: f64 = rhs[1..].trim().parse()?;
        let value = *values
            .get(indicator)
            .ok_or_else(|| anyhow!("indicator {} missing", indicator))?;
        return Ok(value < threshold);
    }
    if let Some(idx) = expr.find('>') {
        let (lhs, rhs) = expr.split_at(idx);
        let indicator = lhs.trim();
        let threshold: f64 = rhs[1..].trim().parse()?;
        let value = *values
            .get(indicator)
            .ok_or_else(|| anyhow!("indicator {} missing", indicator))?;
        return Ok(value > threshold);
    }
    Err(anyhow!("invalid expression"))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn parse_less_than() {
        let mut map = HashMap::new();
        map.insert("rsi".to_string(), 25.0);
        assert!(parse_expression("rsi<30", &map).unwrap());
        assert!(!parse_expression("rsi<20", &map).unwrap());
    }
}

#[derive(Clone, Default)]
pub struct Registry {
    inner: Arc<RwLock<HashMap<String, AutoTradeFilter>>>,
}

impl Registry {
    pub async fn set(&self, principal: String, filter: AutoTradeFilter) {
        let mut guard = self.inner.write().await;
        guard.insert(principal, filter);
    }

    pub async fn clear(&self, principal: &str) {
        let mut guard = self.inner.write().await;
        guard.remove(principal);
    }

    pub async fn get(&self, principal: &str) -> Option<AutoTradeFilter> {
        let guard = self.inner.read().await;
        guard.get(principal).cloned()
    }
}
