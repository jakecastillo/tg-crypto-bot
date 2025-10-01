use std::ffi::{CStr, CString};
use std::os::raw::{c_char, c_double, c_int};
use std::sync::mpsc::{self, Receiver};
use std::sync::Arc;
use std::thread;

use anyhow::Result;
use chrono::{DateTime, Utc};
use ethers::prelude::*;
use futures::StreamExt;
use ta::indicators::{ExponentialMovingAverage, MovingAverageConvergenceDivergence};

#[repr(C)]
pub struct MacdResult {
    pub macd: c_double,
    pub signal: c_double,
    pub histogram: c_double,
    pub error_code: c_int,
}

#[no_mangle]
pub extern "C" fn ta_macd(
    values: *const c_double,
    length: c_int,
    fast: c_int,
    slow: c_int,
    signal: c_int,
) -> MacdResult {
    if values.is_null() || length <= 0 {
        return MacdResult { macd: 0.0, signal: 0.0, histogram: 0.0, error_code: 1 };
    }
    let slice = unsafe { std::slice::from_raw_parts(values, length as usize) };
    let mut ema_fast = ExponentialMovingAverage::new(fast as usize).unwrap();
    let mut ema_slow = ExponentialMovingAverage::new(slow as usize).unwrap();
    let mut macd = MovingAverageConvergenceDivergence::new(fast as usize, slow as usize, signal as usize).unwrap();
    let mut last_macd = 0.0;
    let mut last_signal = 0.0;
    for v in slice {
        ema_fast.next(*v);
        ema_slow.next(*v);
        let macd_val = macd.next(*v);
        last_macd = macd_val.macd;
        last_signal = macd_val.signal;
    }
    let histogram = last_macd - last_signal;
    MacdResult { macd: last_macd, signal: last_signal, histogram, error_code: 0 }
}

#[repr(C)]
pub struct Candle {
    pub open: c_double,
    pub high: c_double,
    pub low: c_double,
    pub close: c_double,
    pub volume: c_double,
    pub timestamp_ms: i64,
    pub pair: *mut c_char,
}

pub struct UniswapHandle {
    receiver: Receiver<Candle>,
}

#[no_mangle]
pub extern "C" fn uniswap_start(pairs: *const c_char) -> *mut UniswapHandle {
    if pairs.is_null() {
        return std::ptr::null_mut();
    }
    let c_str = unsafe { CStr::from_ptr(pairs) };
    let json = match c_str.to_str() {
        Ok(v) => v.to_string(),
        Err(_) => return std::ptr::null_mut(),
    };
    let (tx, rx) = mpsc::channel();
    thread::spawn(move || {
        if let Err(err) = run_uniswap(json, tx.clone()) {
            eprintln!("uniswap thread error: {err:?}");
        }
    });
    Box::into_raw(Box::new(UniswapHandle { receiver: rx }))
}

#[no_mangle]
pub extern "C" fn uniswap_poll(handle: *mut UniswapHandle, out: *mut Candle) -> bool {
    if handle.is_null() || out.is_null() {
        return false;
    }
    let handle = unsafe { &mut *handle };
    match handle.receiver.try_recv() {
        Ok(candle) => {
            unsafe { *out = candle; }
            true
        }
        Err(_) => false,
    }
}

#[no_mangle]
pub extern "C" fn uniswap_stop(handle: *mut UniswapHandle) {
    if handle.is_null() {
        return;
    }
    unsafe { drop(Box::from_raw(handle)); }
}

#[no_mangle]
pub extern "C" fn uniswap_free_string(ptr: *mut c_char) {
    if ptr.is_null() {
        return;
    }
    unsafe { let _ = CString::from_raw(ptr); }
}

fn run_uniswap(json: String, tx: mpsc::Sender<Candle>) -> Result<()> {
    let pairs: Vec<String> = serde_json::from_str(&json)?;
    let runtime = tokio::runtime::Runtime::new()?;
    runtime.block_on(async move {
        let provider = Provider::<Ws>::connect("wss://eth.llamarpc.com").await?;
        let provider = Arc::new(provider);
        for pair in pairs {
            let provider = provider.clone();
            let tx = tx.clone();
            tokio::spawn(async move {
                if let Err(err) = subscribe_pair(provider, pair.clone(), tx.clone()).await {
                    eprintln!("pair subscription failed: {err:?}");
                }
            });
        }
        tokio::signal::ctrl_c().await.ok();
    });
    Ok(())
}

async fn subscribe_pair(provider: Arc<Provider<Ws>>, pair: String, tx: mpsc::Sender<Candle>) -> Result<()> {
    let pool: Address = pair.parse()?;
    let filter = Filter::new().address(pool);
    let mut stream = provider.subscribe_logs(&filter).await?;
    while let Some(log) = stream.next().await {
        let ts: DateTime<Utc> = log.block_timestamp.unwrap_or_else(|| Utc::now());
        let candle = Candle {
            open: 0.0,
            high: 0.0,
            low: 0.0,
            close: 0.0,
            volume: 0.0,
            timestamp_ms: ts.timestamp_millis(),
            pair: CString::new(pair.clone()).unwrap().into_raw(),
        };
        if tx.send(candle).is_err() {
            break;
        }
    }
    Ok(())
}
