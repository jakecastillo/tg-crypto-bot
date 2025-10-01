use anyhow::Result;
use async_trait::async_trait;
use ethers::prelude::*;
use parking_lot::Mutex;
use std::{collections::HashMap, path::PathBuf, sync::Arc};
use thiserror::Error;
use tracing::info;

#[derive(Debug, Error)]
pub enum SignerError {
    #[error("key alias not found: {0}")]
    MissingKey(String),
}

#[async_trait]
pub trait Keystore: Send + Sync {
    async fn sign_transaction(&self, alias: &str, tx: &TypedTransaction) -> Result<Signature>;
    fn address(&self, alias: &str) -> Result<Address>;
}

pub struct MemoryKeystore {
    keys: HashMap<String, Wallet<k256::ecdsa::SigningKey>>,
    nonces: Mutex<HashMap<Address, U256>>,
}

impl MemoryKeystore {
    pub fn from_keys(keys: HashMap<String, Wallet<k256::ecdsa::SigningKey>>) -> Self {
        Self {
            keys,
            nonces: Mutex::new(HashMap::new()),
        }
    }

    fn wallet(&self, alias: &str) -> Result<&Wallet<k256::ecdsa::SigningKey>> {
        self.keys
            .get(alias)
            .ok_or_else(|| SignerError::MissingKey(alias.to_string()).into())
    }
}

#[async_trait]
impl Keystore for MemoryKeystore {
    async fn sign_transaction(&self, alias: &str, tx: &TypedTransaction) -> Result<Signature> {
        let wallet = self.wallet(alias)?.clone();
        let chain_id = wallet.chain_id().unwrap_or(1u64);
        let mut tx = tx.clone();
        tx.set_chain_id(chain_id);

        let address = wallet.address();
        let mut nonces = self.nonces.lock();
        let nonce_entry = nonces.entry(address).or_insert(U256::zero());
        tx.set_nonce(*nonce_entry);
        *nonce_entry += U256::one();
        drop(nonces);

        Ok(wallet.sign_transaction(&tx).await?)
    }

    fn address(&self, alias: &str) -> Result<Address> {
        Ok(self.wallet(alias)?.address())
    }
}

pub struct FileKeystore {
    base: PathBuf,
}

impl FileKeystore {
    pub fn new(path: PathBuf) -> Self {
        Self { base: path }
    }

    pub fn load_wallet(&self, alias: &str) -> Result<Wallet<k256::ecdsa::SigningKey>> {
        let path = self.base.join(format!("{alias}.json"));
        let data = std::fs::read_to_string(&path)?;
        let wallet: LocalWallet = serde_json::from_str(&data)?;
        Ok(wallet)
    }
}

pub async fn sign_eip1559(
    store: Arc<dyn Keystore>,
    alias: &str,
    tx: &TypedTransaction,
) -> Result<Signature> {
    info!(%alias, "signing eip-1559 transaction");
    store.sign_transaction(alias, tx).await
}
