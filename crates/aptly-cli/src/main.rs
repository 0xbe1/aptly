use anyhow::{anyhow, Context, Result};
use aptly_aptos::AptosClient;
use aptly_core::{print_pretty_json, DEFAULT_RPC_URL};
use clap::{Args, Parser, Subcommand};
use serde_json::{json, Value};
use std::io;

#[derive(Parser)]
#[command(name = "aptly-rs")]
#[command(about = "Aptos CLI utilities in Rust")]
struct Cli {
    #[arg(long, global = true, default_value = DEFAULT_RPC_URL)]
    rpc_url: String,

    #[command(subcommand)]
    command: Command,
}

#[derive(Subcommand)]
enum Command {
    Node(NodeCommand),
    Account(AccountCommand),
    Block(BlockCommand),
    Events(EventsCommand),
    Table(TableCommand),
    View(ViewCommand),
    Tx(TxCommand),
    Version,
}

#[derive(Args)]
struct NodeCommand {
    #[command(subcommand)]
    command: NodeSubcommand,
}

#[derive(Subcommand)]
enum NodeSubcommand {
    Ledger,
    Spec,
    Health,
    Info,
    #[command(name = "estimate-gas-price")]
    EstimateGasPrice,
}

#[derive(Args)]
struct AccountCommand {
    #[command(subcommand)]
    command: Option<AccountSubcommand>,
    address: Option<String>,
}

#[derive(Subcommand)]
enum AccountSubcommand {
    Resources(AddressArg),
    Resource(ResourceArgs),
    Modules(AddressArg),
    Module(ModuleArgs),
    Balance(BalanceArgs),
    Txs(TxsArgs),
}

#[derive(Args)]
struct AddressArg {
    address: String,
}

#[derive(Args)]
struct ResourceArgs {
    address: String,
    resource_type: String,
}

#[derive(Args)]
struct ModuleArgs {
    address: String,
    module_name: String,
    #[arg(long)]
    abi: bool,
    #[arg(long)]
    bytecode: bool,
}

#[derive(Args)]
struct BalanceArgs {
    address: String,
    asset_type: Option<String>,
}

#[derive(Args)]
struct TxsArgs {
    address: String,
    #[arg(long, default_value_t = 25)]
    limit: u64,
    #[arg(long, default_value_t = 0)]
    start: u64,
}

#[derive(Args)]
struct BlockCommand {
    #[command(subcommand)]
    command: Option<BlockSubcommand>,
    height: Option<String>,
    #[arg(long, default_value_t = false)]
    with_transactions: bool,
}

#[derive(Subcommand)]
enum BlockSubcommand {
    #[command(name = "by-version")]
    ByVersion(ByVersionArgs),
}

#[derive(Args)]
struct ByVersionArgs {
    version: String,
    #[arg(long, default_value_t = false)]
    with_transactions: bool,
}

#[derive(Args)]
struct EventsCommand {
    address: String,
    creation_number: String,
    #[arg(long, default_value_t = 25)]
    limit: u64,
    #[arg(long, default_value_t = 0)]
    start: u64,
}

#[derive(Args)]
struct TableCommand {
    #[command(subcommand)]
    command: TableSubcommand,
}

#[derive(Subcommand)]
enum TableSubcommand {
    Item(TableItemArgs),
}

#[derive(Args)]
struct TableItemArgs {
    table_handle: String,
    #[arg(long)]
    key_type: String,
    #[arg(long)]
    value_type: String,
    #[arg(long)]
    key: String,
}

#[derive(Args)]
struct ViewCommand {
    function: String,
    #[arg(long = "type-args")]
    type_args: Vec<String>,
    #[arg(long = "args")]
    args: Vec<String>,
}

#[derive(Args)]
struct TxCommand {
    #[command(subcommand)]
    command: Option<TxSubcommand>,
    version_or_hash: Option<String>,
}

#[derive(Subcommand)]
enum TxSubcommand {
    List(TxListArgs),
    Submit,
}

#[derive(Args)]
struct TxListArgs {
    #[arg(long, default_value_t = 25)]
    limit: u64,
    #[arg(long, default_value_t = 0)]
    start: u64,
}

fn main() -> Result<()> {
    let cli = Cli::parse();

    if let Command::Version = cli.command {
        print_version();
        return Ok(());
    }

    let client = AptosClient::new(&cli.rpc_url)?;
    match cli.command {
        Command::Node(command) => run_node(&client, command)?,
        Command::Account(command) => run_account(&client, command)?,
        Command::Block(command) => run_block(&client, command)?,
        Command::Events(command) => run_events(&client, command)?,
        Command::Table(command) => run_table(&client, command)?,
        Command::View(command) => run_view(&client, command)?,
        Command::Tx(command) => run_tx(&client, command)?,
        Command::Version => {}
    }

    Ok(())
}

fn print_version() {
    let version = env!("CARGO_PKG_VERSION");
    let commit_sha = option_env!("APTLY_GIT_SHA").unwrap_or("unknown");
    let build_date = option_env!("APTLY_BUILD_DATE").unwrap_or("unknown");

    println!("aptly-rs {version}");
    println!("commit: {commit_sha}");
    println!("built: {build_date}");
}

fn run_node(client: &AptosClient, command: NodeCommand) -> Result<()> {
    let value = match command.command {
        NodeSubcommand::Ledger => client.get_json("/")?,
        NodeSubcommand::Spec => client.get_json("/spec.json")?,
        NodeSubcommand::Health => client.get_json("/-/healthy")?,
        NodeSubcommand::Info => client.get_json("/info")?,
        NodeSubcommand::EstimateGasPrice => client.get_json("/estimate_gas_price")?,
    };

    print_pretty_json(&value)
}

fn run_account(client: &AptosClient, command: AccountCommand) -> Result<()> {
    match (command.command, command.address) {
        (Some(AccountSubcommand::Resources(args)), _) => {
            let value = client.get_json(&format!("/accounts/{}/resources", args.address))?;
            print_pretty_json(&value)
        }
        (Some(AccountSubcommand::Resource(args)), _) => {
            let encoded = urlencoding::encode(&args.resource_type);
            let value =
                client.get_json(&format!("/accounts/{}/resource/{encoded}", args.address))?;
            print_pretty_json(&value)
        }
        (Some(AccountSubcommand::Modules(args)), _) => {
            let value = client.get_json(&format!("/accounts/{}/modules", args.address))?;
            print_pretty_json(&value)
        }
        (Some(AccountSubcommand::Module(args)), _) => {
            let path = format!("/accounts/{}/module/{}", args.address, args.module_name);
            let value = client.get_json(&path)?;

            if !args.abi && !args.bytecode {
                return print_pretty_json(&value);
            }

            if args.abi {
                let abi = value.get("abi").cloned().unwrap_or(Value::Null);
                return print_pretty_json(&abi);
            }

            let bytecode = value.get("bytecode").cloned().unwrap_or(Value::Null);
            print_pretty_json(&bytecode)
        }
        (Some(AccountSubcommand::Balance(args)), _) => {
            let asset_type = args
                .asset_type
                .unwrap_or_else(|| "0x1::aptos_coin::AptosCoin".to_owned());
            let encoded = urlencoding::encode(&asset_type);
            let value =
                client.get_json(&format!("/accounts/{}/balance/{encoded}", args.address))?;
            print_pretty_json(&value)
        }
        (Some(AccountSubcommand::Txs(args)), _) => {
            let mut path = format!(
                "/accounts/{}/transactions?limit={}",
                args.address, args.limit
            );
            if args.start > 0 {
                path.push_str(&format!("&start={}", args.start));
            }
            let value = client.get_json(&path)?;
            print_pretty_json(&value)
        }
        (None, Some(address)) => {
            let value = client.get_json(&format!("/accounts/{address}"))?;
            print_pretty_json(&value)
        }
        (None, None) => Err(anyhow!("missing address or subcommand")),
    }
}

fn run_block(client: &AptosClient, command: BlockCommand) -> Result<()> {
    match command.command {
        Some(BlockSubcommand::ByVersion(args)) => {
            let path = format!(
                "/blocks/by_version/{}?with_transactions={}",
                args.version, args.with_transactions
            );
            let value = client.get_json(&path)?;
            print_pretty_json(&value)
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
            print_pretty_json(&value)
        }
    }
}

fn run_events(client: &AptosClient, command: EventsCommand) -> Result<()> {
    let mut path = format!(
        "/accounts/{}/events/{}?limit={}",
        command.address, command.creation_number, command.limit
    );
    if command.start > 0 {
        path.push_str(&format!("&start={}", command.start));
    }

    let value = client.get_json(&path)?;
    print_pretty_json(&value)
}

fn run_table(client: &AptosClient, command: TableCommand) -> Result<()> {
    match command.command {
        TableSubcommand::Item(args) => {
            let key_value: Value = serde_json::from_str(&args.key)
                .with_context(|| format!("failed to parse key as JSON: {}", args.key))?;

            let body = json!({
                "key_type": args.key_type,
                "value_type": args.value_type,
                "key": key_value
            });

            let value = client.post_json(&format!("/tables/{}/item", args.table_handle), &body)?;
            print_pretty_json(&value)
        }
    }
}

fn run_view(client: &AptosClient, command: ViewCommand) -> Result<()> {
    let mut parsed_args = Vec::with_capacity(command.args.len());
    for argument in &command.args {
        let parsed: Value = serde_json::from_str(argument)
            .with_context(|| format!("failed to parse argument {argument:?} as JSON"))?;
        parsed_args.push(parsed);
    }

    let body = json!({
        "function": command.function,
        "type_arguments": command.type_args,
        "arguments": parsed_args
    });

    let value = client.post_json("/view", &body)?;
    print_pretty_json(&value)
}

fn run_tx(client: &AptosClient, command: TxCommand) -> Result<()> {
    match (command.command, command.version_or_hash) {
        (Some(TxSubcommand::List(args)), _) => {
            let mut path = format!("/transactions?limit={}", args.limit);
            if args.start > 0 {
                path.push_str(&format!("&start={}", args.start));
            }
            let value = client.get_json(&path)?;
            print_pretty_json(&value)
        }
        (Some(TxSubcommand::Submit), _) => {
            let reader = io::stdin();
            let txn: Value = serde_json::from_reader(reader.lock())
                .context("failed to parse signed transaction JSON from stdin")?;
            let value = client.post_json("/transactions", &txn)?;
            print_pretty_json(&value)
        }
        (None, Some(version_or_hash)) => {
            let path = if version_or_hash.parse::<u64>().is_ok() {
                format!("/transactions/by_version/{version_or_hash}")
            } else {
                format!("/transactions/by_hash/{version_or_hash}")
            };
            let value = client.get_json(&path)?;
            print_pretty_json(&value)
        }
        (None, None) => Err(anyhow!("missing version/hash or subcommand")),
    }
}
