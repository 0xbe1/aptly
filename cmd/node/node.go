package node

import (
	"github.com/spf13/cobra"
)

var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Node information commands",
	Long:  `Commands for querying Aptos node information.`,
}

func init() {
	NodeCmd.AddCommand(ledgerCmd)
	NodeCmd.AddCommand(specCmd)
	NodeCmd.AddCommand(healthCmd)
	NodeCmd.AddCommand(infoCmd)
	NodeCmd.AddCommand(gasPriceCmd)
}
