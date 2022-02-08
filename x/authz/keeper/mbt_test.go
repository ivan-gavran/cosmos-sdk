package keeper_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

type ActionType string

// IG: connect the constant values to the TLA+ spec
const (
	GiveGrant    ActionType = "give grant"
	RevokeGrant  ActionType = "revoke grant"
	ExpireGrant  ActionType = "expire grant"
	ExecuteGrant ActionType = "execute grant"
)

// exec_message": {
// 	"amount": -1,
// 	"message_type": "",
// 	"staking_action": "",
// 	"validator": ""
//   },

type ExecMessageModel struct {
	Amount         int    `json:"amount"`
	Message_Type   string `json:"message_type"`
	Staking_Action string `json:"staking_action"`
	Validator      string `json:"validator"`
}

func GiveSpecifiedGrant(actionTaken interface{}, ctx sdk.Context) {
	// granter := sdk.ValAddress(actionTaken.(map[string]interface{})["granter"].(string))
	// grantee := sdk.ValAddress(actionTaken.(map[string]interface{})["grantee"].(string))
	authorizationType := actionTaken.(map[string]interface{})["authorization_type"].(string)
	fmt.Println(authorizationType)

	switch authorizationType {
	case "stake":
		{
			grantPayload := actionTaken.(map[string]interface{})["grant_payload"]
			allowListIntermediate := grantPayload.(map[string]interface{})["allow_list"]
			allowList := allowListIntermediate.(map[string]interface{})["#set"]
			allowListMap := make(map[string]bool)
			for _, validator := range allowList.([]interface{}) {
				// fmt.Println(validator.(string), reflect.TypeOf(validator.(string)))
				allowListMap[validator.(string)] = true
			}

			fmt.Println(allowListMap, reflect.TypeOf(allowList))
		}
	}
}

func ExecuteAppropriateAction(t *testing.T, state interface{}, ctx sdk.Context) {
	actionTaken := state.(map[string]interface{})["action_taken"]
	actionType := actionTaken.(map[string]interface{})["action_type"].(string)
	switch ActionType(actionType) {
	case GiveGrant:
		GiveSpecifiedGrant(actionTaken, ctx)

	}

}

func TestExecuteItfJson(t *testing.T) {
	app := simapp.Setup(t, false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	traceFileName := "trace_example.itf.json"
	traceFile, err := os.Open(traceFileName)
	if err != nil {
		fmt.Println(err)
	}
	defer traceFile.Close()

	byteValue, _ := ioutil.ReadAll(traceFile)
	var dat map[string]interface{}

	if err := json.Unmarshal(byteValue, &dat); err != nil {
		panic(err)
	}

	states := dat["states"].([]interface{})
	for _, state := range states {
		ExecuteAppropriateAction(t, state, ctx)

	}

	require.True(t, false)
}
