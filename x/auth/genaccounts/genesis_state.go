package genaccounts

import (
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
)

// State to Unmarshal
type GenesisState struct {
	Accounts GenesisAccounts `json:"accounts"`
}

// NewGenesisState creates a new GenesisState object
func NewGenesisState(accounts GenesisAccounts) GenesisState {
	return GenesisState{
		Accounts: accounts,
	}
}

// get the genesis state from the expected app state
func GetGenesisStateFromAppState(cdc *codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState
	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}
	return genesisState
}

// set the genesis state within the expected app state
func SetGenesisStateInAppState(cdc *codec.Codec,
	appState map[string]json.RawMessage, genesisState GenesisState) map[string]json.RawMessage {

	genesisStateBz := cdc.MustMarshalJSON(genesisState)
	appState[ModuleName] = genesisStateBz
	return appState
}

// Sanitize sorts accounts and coin sets.
func (gs GenesisState) Sanitize() {
	sort.Slice(gs.Accounts, func(i, j int) bool {
		return gs.Accounts[i].AccountNumber < gs.Accounts[j].AccountNumber
	})

	for _, acc := range gs.Accounts {
		acc.Coins = acc.Coins.Sort()
	}
}

// ValidateGenesis performs validation of genesis accounts. It
// ensures that there are no duplicate accounts in the genesis state and any
// provided vesting accounts are valid.
func ValidateGenesis(genesisState GenesisState) error {
	addrMap := make(map[string]bool, len(genesisState.Accounts))
	for _, acc := range genesisState.Accounts {
		addrStr := acc.Address.String()

		// disallow any duplicate accounts
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate account found in genesis state; address: %s", addrStr)
		}

		// validate any vesting fields
		if !acc.OriginalVesting.IsZero() {
			if acc.EndTime == 0 {
				return fmt.Errorf("missing end time for vesting account; address: %s", addrStr)
			}

			if acc.StartTime >= acc.EndTime {
				return fmt.Errorf(
					"vesting start time must before end time; address: %s, start: %s, end: %s",
					addrStr,
					time.Unix(acc.StartTime, 0).UTC().Format(time.RFC3339),
					time.Unix(acc.EndTime, 0).UTC().Format(time.RFC3339),
				)
			}
		}

		addrMap[addrStr] = true
	}
	return nil
}
