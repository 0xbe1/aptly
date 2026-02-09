use anyhow::{anyhow, Result};
use aptly_plugin::{discover_move_decompiler, doctor_move_decompiler};
use clap::{Args, Subcommand};

#[derive(Args)]
pub(crate) struct PluginCommand {
    #[command(subcommand)]
    pub(crate) command: PluginSubcommand,
}

#[derive(Subcommand)]
pub(crate) enum PluginSubcommand {
    List,
    Doctor(PluginDoctorArgs),
}

#[derive(Args)]
pub(crate) struct PluginDoctorArgs {
    #[arg(long = "decompiler-bin")]
    pub(crate) decompiler_bin: Option<String>,
}

pub(crate) fn run_plugin(command: PluginCommand) -> Result<()> {
    match command.command {
        PluginSubcommand::List => {
            let plugins = vec![discover_move_decompiler(None)];
            crate::print_serialized(&plugins)
        }
        PluginSubcommand::Doctor(args) => {
            let report = doctor_move_decompiler(args.decompiler_bin.as_deref());
            let ok = report.all_ok();
            crate::print_serialized(&report)?;
            if ok {
                Ok(())
            } else {
                Err(anyhow!(
                    "plugin doctor found issues; see install_hint for remediation"
                ))
            }
        }
    }
}
