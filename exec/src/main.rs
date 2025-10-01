mod config;
mod filters;
mod job;
mod ta_client;

use anyhow::Context;
use hyper::service::{make_service_fn, service_fn};
use hyper::{Body, Response, Server};
use job::Intent;
use metrics_exporter_prometheus::PrometheusBuilder;
use std::net::SocketAddr;
use ta_client::TAClient;
use tokio::time::{sleep, Duration};
use tracing::{error, info};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    tracing_subscriber::fmt::init();
    let cfg = config::Config::load()?;
    info!(?cfg, "starting exec orchestrator");

    let client = redis::Client::open(cfg.redis_url.clone()).context("redis client")?;
    {
        let mut conn = client
            .get_async_connection()
            .await
            .context("redis connection")?;
        ensure_stream_group(&mut conn, &cfg.stream, &cfg.group).await?;
    }

    let metrics_handle = PrometheusBuilder::new().install_recorder()?;
    let metrics_addr: SocketAddr = cfg.metrics_addr.parse()?;
    tokio::spawn(async move {
        let handle = metrics_handle;
        let make_svc = make_service_fn(move |_| {
            let handle = handle.clone();
            async move {
                Ok::<_, hyper::Error>(service_fn(move |_req| {
                    let handle = handle.clone();
                    async move {
                        let body = handle.render();
                        Ok::<_, hyper::Error>(Response::new(Body::from(body)))
                    }
                }))
            }
        });
        if let Err(err) = Server::bind(&metrics_addr).serve(make_svc).await {
            eprintln!("metrics server error: {err}");
        }
    });

    let http_cfg = cfg.clone();
    tokio::spawn(async move {
        serve_health(http_cfg.http_addr).await;
    });

    let conn = client.get_async_connection().await?;
    let ta_client = TAClient::new(cfg.ta_service_url.clone());
    let registry = filters::Registry::default();
    worker_loop(conn, cfg, ta_client, registry).await;
    Ok(())
}

async fn ensure_stream_group(
    conn: &mut redis::aio::Connection,
    stream: &str,
    group: &str,
) -> anyhow::Result<()> {
    let _: Result<(), _> = redis::cmd("XGROUP")
        .arg("CREATE")
        .arg(stream)
        .arg(group)
        .arg("0")
        .arg("MKSTREAM")
        .query_async(conn)
        .await;
    Ok(())
}

async fn worker_loop(
    mut conn: redis::aio::Connection,
    cfg: config::Config,
    ta_client: TAClient,
    registry: filters::Registry,
) {
    loop {
        let resp: redis::Value = match redis::cmd("XREADGROUP")
            .arg("GROUP")
            .arg(&cfg.group)
            .arg(&cfg.consumer)
            .arg("BLOCK")
            .arg(5000)
            .arg("COUNT")
            .arg(1)
            .arg("STREAMS")
            .arg(&cfg.stream)
            .arg(">")
            .query_async(&mut conn)
            .await
        {
            Ok(v) => v,
            Err(err) => {
                error!(%err, "xreadgroup failed");
                sleep(Duration::from_secs(1)).await;
                continue;
            }
        };

        let Some(entries) = parse_stream(resp) else {
            continue;
        };

        for (id, intent) in entries {
            if let Err(err) = handle_intent(&cfg, &intent, &ta_client, &registry).await {
                error!(%err, "failed to handle intent" );
            }
            if let Err(err) = redis::cmd("XACK")
                .arg(&cfg.stream)
                .arg(&cfg.group)
                .arg(&id)
                .query_async(&mut conn)
                .await
            {
                error!(%err, "failed to ack {id}");
            }
        }
    }
}

fn parse_stream(value: redis::Value) -> Option<Vec<(String, Intent)>> {
    match value {
        redis::Value::Bulk(streams) => {
            let mut out = Vec::new();
            for stream in streams {
                if let redis::Value::Bulk(entries) = stream {
                    for entry in entries {
                        if let redis::Value::Bulk(fields) = entry {
                            if fields.len() < 2 {
                                continue;
                            }
                            let id = match &fields[0] {
                                redis::Value::Data(d) => String::from_utf8_lossy(d).to_string(),
                                _ => continue,
                            };
                            if let redis::Value::Bulk(keyvals) = &fields[1] {
                                for chunk in keyvals.chunks(2) {
                                    if chunk.len() != 2 {
                                        continue;
                                    }
                                    if let redis::Value::Data(key) = &chunk[0] {
                                        if key == b"intent" {
                                            if let redis::Value::Data(raw) = &chunk[1] {
                                                if let Ok(intent) =
                                                    serde_json::from_slice::<Intent>(raw)
                                                {
                                                    out.push((id.clone(), intent));
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
            Some(out)
        }
        _ => None,
    }
}

async fn handle_intent(
    cfg: &config::Config,
    intent: &Intent,
    ta_client: &TAClient,
    registry: &filters::Registry,
) -> anyhow::Result<()> {
    let age = intent.age()?;
    info!(intent_id = %intent.id, %age, "processing intent");
    if let Some((action, payload)) = intent.parse_action() {
        handle_action(intent, &action, payload, registry).await
    } else {
        let trade = intent.parse_trade()?;
        handle_trade(cfg, intent, &trade, ta_client, registry).await
    }
}

async fn handle_trade(
    cfg: &config::Config,
    intent: &Intent,
    trade: &job::TradeRequest,
    ta_client: &TAClient,
    registry: &filters::Registry,
) -> anyhow::Result<()> {
    if !trade.force() {
        if let Some(filter) = registry.get(&intent.principal).await {
            let allowed = filter.evaluate(ta_client, &trade.token).await?;
            if !allowed {
                info!(intent_id = %intent.id, expr = %filter.expression, "trade blocked by auto-trade filter");
                return Ok(());
            }
        }
    }
    if cfg.dry_run {
        info!(?trade, "dry run mode - skipping broadcast");
        return Ok(());
    }
    info!(?trade, "executed trade (stub)");
    Ok(())
}

async fn handle_action(
    intent: &Intent,
    action: &str,
    payload: serde_json::Value,
    registry: &filters::Registry,
) -> anyhow::Result<()> {
    match action {
        "set-autotrade-filter" => {
            let enabled = payload
                .get("enabled")
                .and_then(|v| v.as_bool())
                .unwrap_or(true);
            if !enabled {
                registry.clear(&intent.principal).await;
                info!(principal = %intent.principal, "auto-trade filter cleared");
                return Ok(());
            }
            let expression = payload
                .get("expression")
                .and_then(|v| v.as_str())
                .unwrap_or("")
                .to_string();
            let interval = payload
                .get("interval")
                .and_then(|v| v.as_str())
                .unwrap_or("1m")
                .to_string();
            if expression.is_empty() {
                registry.clear(&intent.principal).await;
                return Ok(());
            }
            registry
                .set(
                    intent.principal.clone(),
                    filters::AutoTradeFilter {
                        expression: expression.clone(),
                        interval,
                    },
                )
                .await;
            info!(principal = %intent.principal, expr = %expression, "auto-trade filter set");
            Ok(())
        }
        _ => {
            info!(principal = %intent.principal, action, "action received");
            Ok(())
        }
    }
}

async fn serve_health(addr: String) {
    let make_svc = make_service_fn(|_| async {
        Ok::<_, hyper::Error>(service_fn(|_req| async {
            Ok::<_, hyper::Error>(Response::new(Body::from("ok")))
        }))
    });

    if let Err(err) = Server::bind(&addr.parse().unwrap()).serve(make_svc).await {
        eprintln!("health server error: {err}");
    }
}
