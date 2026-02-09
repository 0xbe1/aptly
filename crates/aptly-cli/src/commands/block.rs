use anyhow::{anyhow, Result};
use aptly_aptos::AptosClient;
use clap::{Args, Subcommand};

#[derive(Args)]
pub(crate) struct BlockCommand {
    #[command(subcommand)]
    pub(crate) command: Option<BlockSubcommand>,
    pub(crate) height: Option<String>,
    #[arg(long, default_value_t = false)]
    pub(crate) with_transactions: bool,
}

#[derive(Subcommand)]
pub(crate) enum BlockSubcommand {
    #[command(name = "by-version")]
    ByVersion(ByVersionArgs),
}

#[derive(Args)]
pub(crate) struct ByVersionArgs {
    pub(crate) version: String,
    #[arg(long, default_value_t = false)]
    pub(crate) with_transactions: bool,
}

pub(crate) fn run_block(client: &AptosClient, command: BlockCommand) -> Result<()> {
    match command.command {
        Some(BlockSubcommand::ByVersion(args)) => {
            let path = format!(
                "/blocks/by_version/{}?with_transactions={}",
                args.version, args.with_transactions
            );
            let value = client.get_json(&path)?;
            crate::print_pretty_json(&value)
        }
        None => {
            let height = command
                .height
                .ok_or_else(|| anyhow!("missing block height or subcommand"))?;
            let path = format!(
                "/blocks/by_height/{height}?with_transactions={}",
                command.with_transactions
            );
            let value = client.get_json(&path)?;
            crate::print_pretty_json(&value)
        }
    }
}
