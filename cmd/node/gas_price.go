package node

import (
	"github.com/0xbe1/aptly/pkg/api"
	"github.com/spf13/cobra"
)

var gasPriceCmd = &cobra.Command{
	Use:   "estimate-gas-price",
	Short: "Estimate current gas price",
	Long:  `Gets the estimated gas price from the Aptos node.`,
	Args:  cobra.NoArgs,
	RunE:  runGasPrice,
}

func runGasPrice(cmd *cobra.Command, args []string) error {
	return api.GetAndPrint(api.BaseURL + "/estimate_gas_price")
}
