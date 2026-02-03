package account

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/0xbe1/aptly/pkg/api"
	"github.com/spf13/cobra"
)

var (
	moduleABI      bool
	moduleBytecode bool
)

var moduleCmd = &cobra.Command{
	Use:   "module <address> <module_name>",
	Short: "Get a specific account module",
	Long: `Fetches and displays a specific module for an account from the Aptos mainnet.

Examples:
  aptly account module 0x1 coin
  aptly account module 0x1 coin --abi
  aptly account module 0x1 coin --bytecode`,
	Args: cobra.ExactArgs(2),
	RunE: runModule,
}

func init() {
	moduleCmd.Flags().BoolVar(&moduleABI, "abi", false, "Output only the ABI")
	moduleCmd.Flags().BoolVar(&moduleBytecode, "bytecode", false, "Output only the bytecode")
}

func runModule(cmd *cobra.Command, args []string) error {
	// If no filter flags, use the simple path
	if !moduleABI && !moduleBytecode {
		return api.GetAndPrint(fmt.Sprintf("%s/accounts/%s/module/%s", api.BaseURL, args[0], args[1]))
	}

	// Fetch the module data
	url := fmt.Sprintf("%s/accounts/%s/module/%s", api.BaseURL, args[0], args[1])
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var module map[string]any
	if err := json.Unmarshal(body, &module); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Output filtered result
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")

	if moduleABI {
		return encoder.Encode(module["abi"])
	}
	return encoder.Encode(module["bytecode"])
}
