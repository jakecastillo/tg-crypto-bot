use anyhow::Result;
use async_trait::async_trait;
use solana_client::nonblocking::rpc_client::RpcClient;
use solana_sdk::{signature::Keypair, transaction::Transaction};
use thiserror::Error;
use tracing::info;

#[derive(Debug, Error)]
pub enum SolanaConnectorError {
    #[error("jito bundle submission not yet implemented")]
    JitoUnsupported,
}

#[async_trait]
pub trait SolanaExecutor {
    async fn submit_bundle(&self, txs: Vec<Transaction>) -> Result<()>;
    async fn place_spot_order(&self, market: String, size: u64) -> Result<()>;
}

pub struct JitoPlaceholder {
    rpc: RpcClient,
}

impl JitoPlaceholder {
    pub fn new(url: &str) -> Self {
        Self {
            rpc: RpcClient::new(url.to_string()),
        }
    }
}

#[async_trait]
impl SolanaExecutor for JitoPlaceholder {
    async fn submit_bundle(&self, _txs: Vec<Transaction>) -> Result<()> {
        Err(SolanaConnectorError::JitoUnsupported.into())
    }

    async fn place_spot_order(&self, market: String, size: u64) -> Result<()> {
        info!(%market, %size, "simulating solana spot order");
        Ok(())
    }
}
