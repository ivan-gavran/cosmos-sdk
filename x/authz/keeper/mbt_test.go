package keeper_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta2"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

var app *simapp.SimApp
var ctx sdk.Context

const GRANT_SUCCESS string = "grant_success"
const GRANT_FAILED string = "grant_failed"
const REVOKE_SUCCESS string = "revoke_success"
const REVOKE_FAILED string = "revoke_failed"

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
	Amount       int    `json:"amount"`
	Message_Type string `json:"message_type"`
	Validator    string `json:"validator"`
}

type ValidatorList struct {
	Validators []string `json:"#set"`
}

type GrantPayloadModel struct {
	Limit              int           `json:"limit"`
	SpecialValue       string        `json:"special_value"`
	AllowList          ValidatorList `json:"allow_list"`
	DenyList           ValidatorList `json:"deny_list"`
	AuthorizationLogic string        `json:"authorization_logic"`
}

type GrantModel struct {
	Granter        string `json:"granter"`
	Grantee        string `json:"grantee"`
	SdkMessageType string `json:"sdk_message_type"`
}

type ActionModel struct {
	ActionType   string            `json:"action_type"`
	Grant        GrantModel        `json:"grant"`
	ExecMessage  ExecMessageModel  `json:"exec_message"`
	GrantPayload GrantPayloadModel `json:"grant_payload"`
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

func getValAddrList(names []string) []sdk.ValAddress {
	addrList := make([]sdk.ValAddress, len(names))
	for i, name := range names {
		addrList[i] = sdk.ValAddress(name)
	}

	return addrList
}

func stakeMsgTlaToCosmosURL(msg_type string) (string, error) {
	switch msg_type {
	case "msg_delegate":
		return sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}), nil
	case "msg_undelegate":
		return sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}), nil
	case "msg_redelegate":
		return sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}), nil
	default:
		return "", sdkerrors.ErrInvalidType.Wrapf("unknown authorization for message type %T", msg_type)
	}
}

func genericMsgTlaToCosmosURL(msg_type string) (string, error) {

	switch msg_type {
	case "msg_delegate":
		return sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}), nil
	case "msg_undelegate":
		return sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}), nil
	case "msg_redelegate":
		return sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}), nil
	case "msg_send":
		return sdk.MsgTypeURL(&banktypes.MsgSend{}), nil
	case "msg_alpha":
		return sdk.MsgTypeURL(&govtypes.MsgVote{}), nil
	default:
		return "", fmt.Errorf("invalid message")
	}
}

func giveSpecifiedGrant(actionTaken ActionModel, outcome string, t *testing.T) {
	grantPayload := actionTaken.GrantPayload
	var authorization authz.Authorization

	switch grantPayload.AuthorizationLogic {
	case "stake":
		{

			allowList := getValAddrList(grantPayload.AllowList.Validators)
			denyList := getValAddrList(grantPayload.DenyList.Validators)

			var stakingAuthorizationType stakingtypes.AuthorizationType
			switch actionTaken.Grant.SdkMessageType {
			case "msg_delegate":
				stakingAuthorizationType = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE
			case "msg_undelegate":
				stakingAuthorizationType = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE
			case "msg_redelegate":
				stakingAuthorizationType = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE
			default:
				stakingAuthorizationType = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNSPECIFIED
			}

			var limitAddr *sdk.Coin
			if grantPayload.SpecialValue == "inf" {
				limitAddr = nil
			} else {
				limit := sdk.NewInt64Coin("testCoin", int64(grantPayload.Limit))
				limitAddr = &limit
			}
			var err error
			authorization, err = stakingtypes.NewStakeAuthorization(
				allowList,
				denyList,
				stakingAuthorizationType,
				limitAddr)
			// IG: is there a benefit in adding these low-level checks?
			require.NoError(t, err)
			authorizationMsgTypeURL, err := genericMsgTlaToCosmosURL(actionTaken.Grant.SdkMessageType)
			require.NoError(t, err)
			require.Equal(t, authorization.MsgTypeURL(), authorizationMsgTypeURL)

		}
	case "send":
		{

			var limit sdk.Coins
			limit = sdk.NewCoins(sdk.NewCoin("testCoin", sdk.NewInt(int64(grantPayload.Limit))))
			authorization = banktypes.NewSendAuthorization(limit)

		}
	case "generic":
		{

			msg_url, err := genericMsgTlaToCosmosURL(actionTaken.Grant.SdkMessageType)
			require.NoError(t, err)
			authorization = authz.NewGenericAuthorization(msg_url)
		}
	default:
		sdkerrors.ErrInvalidType.Wrapf("modelator: Invalid type for the authorization logic")

	}
	require.NoError(t, authorization.ValidateBasic())

	granterAddr := sdk.AccAddress(actionTaken.Grant.Granter)
	granteeAddr := sdk.AccAddress(actionTaken.Grant.Grantee)
	now := ctx.BlockHeader().Time
	err := app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, authorization, now.Add(time.Hour))
	require.NoError(t, err)
	msgAuthorizationType := authorization.MsgTypeURL()

	storedAuthz, _ := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, msgAuthorizationType)

	if outcome == GRANT_SUCCESS {
		require.NotNil(t, storedAuthz)
	} else if outcome == GRANT_FAILED {
		require.Nil(t, storedAuthz)
	} else {
		// IG: what is the best way to handle these parts, where something unexpected happened inside the driver?
		require.True(t, false)
	}

}

func revokeSpecifiedGrant(actionTaken ActionModel, outcome string, t *testing.T) {
	granterAddr := sdk.AccAddress(actionTaken.Grant.Granter)
	granteeAddr := sdk.AccAddress(actionTaken.Grant.Grantee)
	authorizationType, err := genericMsgTlaToCosmosURL(actionTaken.Grant.SdkMessageType)
	require.NoError(t, err)
	err = app.AuthzKeeper.DeleteGrant(ctx, granteeAddr, granterAddr, authorizationType)
	require.NoError(t, err)

	authorization, _ := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, authorizationType)
	if outcome == REVOKE_SUCCESS {
		require.Nil(t, authorization)
	} else if outcome == REVOKE_FAILED {
		require.NotNil(t, authorization)
	} else {
		require.True(t, false)
	}

}

func expireSpecifiedGrant(actionTaken ActionModel, outcome string, t *testing.T) {
	granterAddr := sdk.AccAddress(actionTaken.Grant.Granter)
	granteeAddr := sdk.AccAddress(actionTaken.Grant.Grantee)
	authorizationType, err := genericMsgTlaToCosmosURL(actionTaken.Grant.SdkMessageType)
	require.NoError(t, err)
	authorization, _ := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, authorizationType)
	now := ctx.BlockHeader().Time
	// expiring a grant is done by updating the previous authorization by a new one that is in the past
	err = app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, authorization, now.Add(-1*time.Hour))
	require.NoError(t, err)

}

func executeAppropriateAction(state StateModel, t *testing.T) {
	actionTaken := state.ActionTaken
	outcome := state.OutcomeStatus
	fmt.Println(actionTaken.ActionType)
	switch ActionType(actionTaken.ActionType) {
	case GiveGrant:
		giveSpecifiedGrant(actionTaken, outcome, t)
	case RevokeGrant:
		revokeSpecifiedGrant(actionTaken, outcome, t)
	case ExpireGrant:
		expireSpecifiedGrant(actionTaken, outcome, t)
	default:
		fmt.Printf("hi, nothing here for action type %s\n", actionTaken.ActionType)
	}

}

func TestExecuteItfJson(t *testing.T) {
	app = simapp.Setup(t, false)
	ctx = app.BaseApp.NewContext(false, tmproto.Header{})
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
		executeAppropriateAction(state, t)
	}

	// TODO: in the last step, check the full state of active grants and their authorizations

	require.True(t, false)
}
