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

type ValidatorList struct {
	Validators []string `json:"set"`
}

type GrantPayloadModel struct {
	Limit        int           `json:"limit"`
	SpecialValue string        `json:"special_value"`
	StakingType  string        `json:"staking_type"`
	AllowList    ValidatorList `json:"allow_list"`
	DenyList     ValidatorList `json:"deny_list"`
}

type ActionModel struct {
	ActionType        string            `json:"action_type"`
	AuthorizationType string            `json:"authorization_type"`
	ExecMessage       ExecMessageModel  `json:"exec_message"`
	GrantPayload      GrantPayloadModel `json:"grant_payload"`
	Granter           string            `json:"granter"`
	Grantee           string            `json:"grantee"`
}

type StateModel struct {
	Meta          interface{} `json:"#meta"`
	ActionTaken   ActionModel `json:"action_taken"`
	ActiveGrants  interface{} `json:"active_grants"`
	ExpiredGrants interface{} `json:"expired_grants"`
	NumExecs      int         `json:"num_execs"`
	NumGrants     int         `json:"num_grants"`
	OutcomeStatus string      `json:"outcome_status"`
}

type MainJsonStruct struct {
	Meta   interface{}   `json:"#meta"`
	Vars   []interface{} `json:"vars"`
	States []StateModel  `json:"states"`
}

func GiveSpecifiedGrant(actionTaken ActionModel, ctx sdk.Context) {
	// granter := sdk.ValAddress(actionTaken.(map[string]interface{})["granter"].(string))
	// grantee := sdk.ValAddress(actionTaken.(map[string]interface{})["grantee"].(string))
	authorizationType := actionTaken.AuthorizationType
	fmt.Println(authorizationType)

	switch authorizationType {
	case "stake":
		{
			grantPayload := actionTaken.GrantPayload
			allowList := grantPayload.AllowList.Validators
			denyList := grantPayload.DenyList.Validators
			fmt.Println(allowList, reflect.TypeOf(allowList))
		}
	}
}

func ExecuteAppropriateAction(state StateModel, ctx sdk.Context, t *testing.T) {
	actionTaken := state.ActionTaken

	switch ActionType(actionTaken.ActionType) {
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
	var jsonData MainJsonStruct
	json.Unmarshal(byteValue, &jsonData)
	for _, state := range jsonData.States {
		ExecuteAppropriateAction(state, ctx, t)
	}

	require.True(t, false)
}
