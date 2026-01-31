package tx

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/api"
	"github.com/spf13/cobra"
)

var transfersCmd = &cobra.Command{
	Use:   "transfers <tx_version>",
	Short: "Show asset transfers in a transaction",
	Long:  `Lists Withdraw/Deposit events from a transaction.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runTransfers,
}

type TransferEvent struct {
	Type          string `json:"type"` // "withdraw" or "deposit"
	Account       string `json:"account"`
	FungibleStore string `json:"fungible_store"`
	Asset         string `json:"asset"`
	Amount        string `json:"amount"`
}

func runTransfers(cmd *cobra.Command, args []string) error {
	version, err := strconv.ParseUint(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid transaction version: %w", err)
	}

	client, err := aptos.NewClient(aptos.MainnetConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	tx, err := client.TransactionByVersion(version)
	if err != nil {
		return fmt.Errorf("failed to fetch transaction %d: %w", version, err)
	}

	// Extract store info from tx changes
	storeInfo := extractTransferStoreInfo(tx)

	// Extract flow events
	events := extractTransferEvents(tx, storeInfo, client, version)

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(events); err != nil {
		return fmt.Errorf("failed to encode response: %w", err)
	}

	return nil
}

type transferStoreMetadata struct {
	owner string
	asset string
}

func extractTransferStoreInfo(tx *api.CommittedTransaction) map[string]transferStoreMetadata {
	info := make(map[string]transferStoreMetadata)

	data, err := json.Marshal(tx)
	if err != nil {
		return info
	}

	var txMap map[string]any
	if err := json.Unmarshal(data, &txMap); err != nil {
		return info
	}

	inner, ok := txMap["Inner"].(map[string]any)
	if !ok {
		return info
	}

	changes, ok := inner["Changes"].([]any)
	if !ok {
		return info
	}

	// Extract owners from ObjectCore
	owners := make(map[string]string)
	for _, change := range changes {
		changeMap, ok := change.(map[string]any)
		if !ok {
			continue
		}
		if changeMap["Type"] != "write_resource" {
			continue
		}
		changeInner, ok := changeMap["Inner"].(map[string]any)
		if !ok {
			continue
		}
		resourceData, ok := changeInner["data"].(map[string]any)
		if !ok {
			continue
		}
		if resourceData["type"] != "0x1::object::ObjectCore" {
			continue
		}
		address, _ := changeInner["address"].(string)
		data, ok := resourceData["data"].(map[string]any)
		if !ok {
			continue
		}
		owner, _ := data["owner"].(string)
		owners[address] = owner
	}

	// Extract assets from FungibleStore
	for _, change := range changes {
		changeMap, ok := change.(map[string]any)
		if !ok {
			continue
		}
		if changeMap["Type"] != "write_resource" {
			continue
		}
		changeInner, ok := changeMap["Inner"].(map[string]any)
		if !ok {
			continue
		}
		resourceData, ok := changeInner["data"].(map[string]any)
		if !ok {
			continue
		}
		resourceType, _ := resourceData["type"].(string)
		if resourceType != "0x1::fungible_asset::FungibleStore" {
			continue
		}
		address, _ := changeInner["address"].(string)
		data, ok := resourceData["data"].(map[string]any)
		if !ok {
			continue
		}
		metadata, ok := data["metadata"].(map[string]any)
		if !ok {
			continue
		}
		asset, _ := metadata["inner"].(string)

		info[address] = transferStoreMetadata{
			owner: owners[address],
			asset: asset,
		}
	}

	return info
}

func extractTransferEvents(tx *api.CommittedTransaction, storeInfo map[string]transferStoreMetadata, client *aptos.Client, version uint64) []TransferEvent {
	var events []TransferEvent

	data, err := json.Marshal(tx)
	if err != nil {
		return events
	}

	var txMap map[string]any
	if err := json.Unmarshal(data, &txMap); err != nil {
		return events
	}

	inner, ok := txMap["Inner"].(map[string]any)
	if !ok {
		return events
	}

	txEvents, ok := inner["Events"].([]any)
	if !ok {
		return events
	}

	for _, event := range txEvents {
		eventMap, ok := event.(map[string]any)
		if !ok {
			continue
		}

		eventType, _ := eventMap["Type"].(string)
		var transferType string
		switch eventType {
		case "0x1::fungible_asset::Withdraw":
			transferType = "withdraw"
		case "0x1::fungible_asset::Deposit":
			transferType = "deposit"
		default:
			continue
		}

		eventData, ok := eventMap["Data"].(map[string]any)
		if !ok {
			continue
		}

		store, _ := eventData["store"].(string)
		amount, _ := eventData["amount"].(string)

		meta, ok := storeInfo[store]
		if !ok {
			meta = queryTransferStoreInfo(client, store, version)
		}

		events = append(events, TransferEvent{
			Type:          transferType,
			Account:       meta.owner,
			FungibleStore: store,
			Asset:         meta.asset,
			Amount:        amount,
		})
	}

	return events
}

func queryTransferStoreInfo(client *aptos.Client, store string, version uint64) transferStoreMetadata {
	meta := transferStoreMetadata{}

	addr := aptos.AccountAddress{}
	if err := addr.ParseStringRelaxed(store); err != nil {
		return meta
	}

	objCore, err := client.AccountResource(addr, "0x1::object::ObjectCore", version)
	if err == nil {
		if data, ok := objCore["data"].(map[string]any); ok {
			if owner, ok := data["owner"].(string); ok {
				meta.owner = owner
			}
		}
	}

	fsResource, err := client.AccountResource(addr, "0x1::fungible_asset::FungibleStore", version)
	if err == nil {
		if data, ok := fsResource["data"].(map[string]any); ok {
			if metadata, ok := data["metadata"].(map[string]any); ok {
				if asset, ok := metadata["inner"].(string); ok {
					meta.asset = asset
				}
			}
		}
	}

	return meta
}
