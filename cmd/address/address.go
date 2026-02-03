package address

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const labelsURL = "https://raw.githubusercontent.com/ThalaLabs/aptos-labels/main/mainnet.json"

var AddressCmd = &cobra.Command{
	Use:   "address <query>",
	Short: "Search for address labels",
	Long:  `Searches the Aptos address labels from ThalaLabs for case-insensitive matches.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAddress,
}

func runAddress(cmd *cobra.Command, args []string) error {
	query := strings.ToLower(args[0])

	resp, err := http.Get(labelsURL)
	if err != nil {
		return fmt.Errorf("failed to fetch labels: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var labels map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&labels); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	matches := make(map[string]string)
	for addr, label := range labels {
		if strings.Contains(strings.ToLower(label), query) {
			matches[addr] = label
		}
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(matches)
}
