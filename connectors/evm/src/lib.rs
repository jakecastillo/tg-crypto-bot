use anyhow::Result;
use async_trait::async_trait;
use ethers::prelude::*;
use std::sync::Arc;
use thiserror::Error;
use tracing::info;

#[derive(Debug, Error)]
pub enum EvmConnectorError {
    #[error("token not safelisted: {0}")]
    TokenNotSafelisted(String),
    #[error("router not supported: {0}")]
    RouterNotSupported(String),
    #[error("provider error: {0}")]
    Provider(String),
}

#[async_trait]
pub trait DexConnector {
    async fn quote(&self, token_in: Address, token_out: Address, amount_in: U256) -> Result<U256>;
    async fn execute_swap(
        &self,
        wallet: Arc<SignerMiddleware<Provider<Ws>, Wallet<k256::ecdsa::SigningKey>>>,
        params: SwapParams,
    ) -> Result<TxHash>;
}

#[derive(Debug, Clone)]
pub struct SwapParams {
    pub router: Address,
    pub token_in: Address,
    pub token_out: Address,
    pub amount_in: U256,
    pub min_out: U256,
    pub deadline: U256,
}

pub struct UniswapV2Connector {
    provider: Provider<Ws>,
    safelist: Vec<Address>,
}

impl UniswapV2Connector {
    pub async fn new(ws_url: &str, safelist: Vec<Address>) -> Result<Self> {
        let provider = Provider::<Ws>::connect(ws_url).await?;
        Ok(Self { provider, safelist })
    }

    fn ensure_safelisted(&self, token: Address) -> Result<()> {
        if self.safelist.contains(&token) {
            Ok(())
        } else {
            Err(EvmConnectorError::TokenNotSafelisted(format!("0x{:x}", token)).into())
        }
    }
}

#[async_trait]
impl DexConnector for UniswapV2Connector {
    async fn quote(&self, token_in: Address, token_out: Address, amount_in: U256) -> Result<U256> {
        self.ensure_safelisted(token_in)?;
        self.ensure_safelisted(token_out)?;
        // Placeholder for router call; use getAmountsOut in production
        info!("quoting swap", ?token_in, ?token_out, ?amount_in);
        Ok(amount_in)
    }

    async fn execute_swap(
        &self,
        wallet: Arc<SignerMiddleware<Provider<Ws>, Wallet<k256::ecdsa::SigningKey>>>,
        params: SwapParams,
    ) -> Result<TxHash> {
        self.ensure_safelisted(params.token_in)?;
        self.ensure_safelisted(params.token_out)?;

        let tx = TransactionRequest::new()
            .to(params.router)
            .value(U256::zero())
            .data(Bytes::from_static(b""))
            .from(wallet.address());

        let pending = wallet.send_transaction(tx, None).await?;
        let receipt = pending.await?;
        Ok(receipt.transaction_hash)
    }
}
