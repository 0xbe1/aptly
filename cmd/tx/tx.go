package tx

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

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
	TxCmd.AddCommand(simulateCmd)
}

func runTx(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}

	rawJSON, err := fetchTransactionRaw(args[0])
	if err != nil {
		return err
	}

	var data any
	if err := json.Unmarshal(rawJSON, &data); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}

// fetchTransactionRaw fetches raw JSON from the Aptos API
func fetchTransactionRaw(versionOrHash string) ([]byte, error) {
	var url string
	if _, err := strconv.ParseUint(versionOrHash, 10, 64); err == nil {
		url = fmt.Sprintf("https://api.mainnet.aptoslabs.com/v1/transactions/by_version/%s", versionOrHash)
	} else {
		url = fmt.Sprintf("https://api.mainnet.aptoslabs.com/v1/transactions/by_hash/%s", versionOrHash)
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}
