use anyhow::Result;
use serde_json::Value;

pub const DEFAULT_RPC_URL: &str = "https://api.mainnet.aptoslabs.com/v1";

pub fn print_pretty_json(value: &Value) -> Result<()> {
    let rendered = serde_json::to_string_pretty(value)?;
    println!("{rendered}");
    Ok(())
}
