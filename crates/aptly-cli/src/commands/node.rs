use anyhow::Result;
use aptly_aptos::AptosClient;
use clap::{Args, Subcommand};

#[derive(Args)]
pub(crate) struct NodeCommand {
    #[command(subcommand)]
    pub(crate) command: NodeSubcommand,
}

#[derive(Subcommand)]
pub(crate) enum NodeSubcommand {
    Ledger,
    Spec,
    Health,
    Info,
    #[command(name = "estimate-gas-price")]
    EstimateGasPrice,
}

pub(crate) fn run_node(client: &AptosClient, command: NodeCommand) -> Result<()> {
    let value = match command.command {
        NodeSubcommand::Ledger => client.get_json("/")?,
        NodeSubcommand::Spec => client.get_json("/spec.json")?,
        NodeSubcommand::Health => client.get_json("/-/healthy")?,
        NodeSubcommand::Info => client.get_json("/info")?,
        NodeSubcommand::EstimateGasPrice => client.get_json("/estimate_gas_price")?,
    };

    crate::print_pretty_json(&value)
}
