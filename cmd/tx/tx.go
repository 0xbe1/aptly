package tx

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/spf13/cobra"
)

var TxCmd = &cobra.Command{
	Use:   "tx <version_or_hash>",
	Short: "Transaction commands",
	Long:  `View and analyze Aptos transactions. Run with a version or hash to view the raw transaction.`,
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTx,
}

func init() {
	TxCmd.AddCommand(balanceChangeCmd)
	TxCmd.AddCommand(transfersCmd)
	TxCmd.AddCommand(traceCmd)
	TxCmd.AddCommand(graphCmd)
}

func runTx(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	client, err := aptos.NewClient(aptos.MainnetConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	userTx, _, err := fetchTransaction(client, args[0])
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(userTx); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}
