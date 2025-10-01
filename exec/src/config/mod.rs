use serde::Deserialize;

#[derive(Debug, Deserialize, Clone)]
pub struct Config {
    pub redis_url: String,
    pub stream: String,
    pub group: String,
    pub consumer: String,
    pub http_addr: String,
    pub metrics_addr: String,
    pub dry_run: bool,
    pub ta_service_url: String,
}

impl Default for Config {
    fn default() -> Self {
        Self {
            redis_url: "redis://127.0.0.1:6379".to_string(),
            stream: "trade-intents".to_string(),
            group: "exec".to_string(),
            consumer: format!("exec-{}", std::process::id()),
            http_addr: "0.0.0.0:8081".to_string(),
            metrics_addr: "0.0.0.0:9101".to_string(),
            dry_run: true,
            ta_service_url: "http://ta-service:9100".to_string(),
        }
    }
}

impl Config {
    pub fn load() -> anyhow::Result<Self> {
        let mut settings = config::Config::builder()
            .add_source(config::Environment::with_prefix("TG_TRADER_EXEC").separator("__"))
            .build()?;

        let cfg: Config = settings.try_deserialize().unwrap_or_default();
        Ok(cfg)
    }
}
