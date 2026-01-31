package tx

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/spf13/cobra"
)

var prettyOutput bool

var graphCmd = &cobra.Command{
	Use:   "graph <version_or_hash>",
	Short: "Show asset transfers as a graph",
	Long:  `Pairs withdraw→deposit events for the same asset to show transfer flows.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runGraph,
}

func init() {
	graphCmd.Flags().BoolVar(&prettyOutput, "pretty", false, "Human-readable output")
}

type Transfer struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Asset  string `json:"asset"`
	Amount string `json:"amount"`
}

type OrphanEvent struct {
	Account string `json:"account"`
	Asset   string `json:"asset"`
	Amount  string `json:"amount"`
}

type Orphans struct {
	In  []OrphanEvent `json:"in"`
	Out []OrphanEvent `json:"out"`
}

type TransferGraph struct {
	Transfers []Transfer `json:"transfers"`
	Orphans   Orphans    `json:"orphans"`
}

func runGraph(cmd *cobra.Command, args []string) error {
	client, err := aptos.NewClient(aptos.MainnetConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	userTx, version, err := fetchTransaction(client, args[0])
	if err != nil {
		return err
	}

	storeInfo := extractTransferStoreInfoFromUserTx(userTx)
	graph := buildTransferGraph(userTx, storeInfo, client, version)

	if prettyOutput {
		printPrettyGraph(graph)
		return nil
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(graph); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}

type pendingWithdraw struct {
	account string
	amount  string
}

func buildTransferGraph(userTx *api.UserTransaction, storeInfo map[string]transferStoreMetadata, client *aptos.Client, version uint64) TransferGraph {
	graph := TransferGraph{
		Transfers: []Transfer{},
		Orphans: Orphans{
			In:  []OrphanEvent{},
			Out: []OrphanEvent{},
		},
	}

	// Track pending withdraws by asset: asset -> list of pending withdraws
	pendingWithdraws := make(map[string][]pendingWithdraw)

	for _, event := range userTx.Events {
		store := getString(event.Data, "store")
		amount := getString(event.Data, "amount")

		meta, ok := storeInfo[store]
		if !ok {
			meta = queryTransferStoreInfo(client, store, version)
		}

		switch event.Type {
		case "0x1::fungible_asset::Withdraw":
			pendingWithdraws[meta.asset] = append(pendingWithdraws[meta.asset], pendingWithdraw{
				account: meta.owner,
				amount:  amount,
			})

		case "0x1::fungible_asset::Deposit":
			pending := pendingWithdraws[meta.asset]
			if len(pending) > 0 {
				// Match with first pending withdraw for this asset
				withdraw := pending[0]
				pendingWithdraws[meta.asset] = pending[1:]

				graph.Transfers = append(graph.Transfers, Transfer{
					From:   withdraw.account,
					To:     meta.owner,
					Asset:  meta.asset,
					Amount: amount,
				})
			} else {
				// No matching withdraw - orphan in
				graph.Orphans.In = append(graph.Orphans.In, OrphanEvent{
					Account: meta.owner,
					Asset:   meta.asset,
					Amount:  amount,
				})
			}
		}
	}

	// Any remaining pending withdraws are orphan outs
	for asset, pending := range pendingWithdraws {
		for _, w := range pending {
			graph.Orphans.Out = append(graph.Orphans.Out, OrphanEvent{
				Account: w.account,
				Asset:   asset,
				Amount:  w.amount,
			})
		}
	}

	return graph
}

func printPrettyGraph(graph TransferGraph) {
	// Group transfers by sender
	bySender := make(map[string][]Transfer)
	for _, t := range graph.Transfers {
		bySender[t.From] = append(bySender[t.From], t)
	}

	for sender, transfers := range bySender {
		fmt.Println(truncateAddress(sender))
		for _, t := range transfers {
			fmt.Printf("  → %s   %s %s\n", truncateAddress(t.To), t.Amount, truncateAddress(t.Asset))
		}
		fmt.Println()
	}

	if len(graph.Orphans.In) > 0 || len(graph.Orphans.Out) > 0 {
		fmt.Println("Orphans:")
		for _, o := range graph.Orphans.In {
			fmt.Printf("  IN:  %s  %s %s\n", truncateAddress(o.Account), o.Amount, truncateAddress(o.Asset))
		}
		for _, o := range graph.Orphans.Out {
			fmt.Printf("  OUT: %s  %s %s\n", truncateAddress(o.Account), o.Amount, truncateAddress(o.Asset))
		}
	}
}

func truncateAddress(addr string) string {
	if len(addr) <= 12 {
		return addr
	}
	return addr[:6] + ".." + addr[len(addr)-4:]
}
