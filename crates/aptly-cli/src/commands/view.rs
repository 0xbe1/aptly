use anyhow::{Context, Result};
use aptly_aptos::AptosClient;
use clap::Args;
use serde_json::{json, Value};

#[derive(Args)]
pub(crate) struct ViewCommand {
    pub(crate) function: String,
    #[arg(long = "type-args")]
    pub(crate) type_args: Vec<String>,
    #[arg(long = "args")]
    pub(crate) args: Vec<String>,
}

pub(crate) fn run_view(client: &AptosClient, command: ViewCommand) -> Result<()> {
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
    crate::print_pretty_json(&value)
}
