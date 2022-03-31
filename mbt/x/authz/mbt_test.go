package mbt

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"
)

var app *simapp.SimApp
var ctx sdk.Context
var accountNameToAddress map[string]sdk.AccAddress
var validators map[string]sdk.ValAddress
var PKs []types.PubKey

var modelMsgToCosmosURL map[string]string
var cosmosURLToModelMsg map[string]string

const GRANT_SUCCESS string = "grant_success"
const GRANT_FAILED string = "grant_failed"
const REVOKE_SUCCESS string = "revoke_success"
const REVOKE_FAILED string = "revoke_failed"

const NONEXISTENT_GRANT_EXEC string = "non_existent_auth"
const EXPIRED_AUTH_EXEC string = "tried to execute an expired grant"
const INSUFFICIENT_GRANT_EXEC string = "insufficient_auth_exec"
const SUCCESSFUL_AUTH_EXEC string = "successful_auth_exec"
const INAPPROPRIATE_AUTH_STAKE_NOT_ALLOW string = "inappropriate_auth_stake_not_allow"
const INAPPROPRIATE_AUTH_STAKE_DENY string = "inappropriate_auth_stake_deny"
const INAPPROPRIATE_AUTH_FOR_MESSAGE string = "message_not_supported_by_the_authorization"
const GRANT_SPENT string = "grant_spent"

type ActionType string

// IG: connect the constant values to the TLA+ spec
const (
	GiveGrant    ActionType = "give grant"
	RevokeGrant  ActionType = "revoke grant"
	ExpireGrant  ActionType = "expire grant"
	ExecuteGrant ActionType = "execute grant"
)

type MessageType string

const (
	MsgDelegate   MessageType = "msg_delegate"
	MsgUndelegate MessageType = "msg_undelegate"
	MsgRedelegate MessageType = "msg_redelegate"
	MsgSend       MessageType = "msg_send"
	MsgOther      MessageType = "msg_alpha"
)

type AuthorizationLogic string

const (
	Generic AuthorizationLogic = "generic"
	Send    AuthorizationLogic = "send"
	Stake   AuthorizationLogic = "stake"
)

type ExecMessageModel struct {
	Amount       int    `json:"amount"`
	Message_Type string `json:"message_type"`
	Validator    string `json:"validator"`
	NewValidator string `json:"new_validator"`
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

type ExecOutcomeModel struct {
	Accept      bool              `json:"accept"`
	Delete      bool              `json:"delete"`
	Description string            `json:"description"`
	Updated     GrantPayloadModel `json:"updated"`
}

type ActionModel struct {
	ActionType   string            `json:"action_type"`
	Grant        GrantModel        `json:"grant"`
	ExecMessage  ExecMessageModel  `json:"exec_message"`
	ExecOutcome  ExecOutcomeModel  `json:"exec_outcome"`
	GrantPayload GrantPayloadModel `json:"grant_payload"`
}

type ActiveGrantsModel struct {
	Grantee            string   `json:"grantee,omitempty"`
	Granter            string   `json:"granter,omitempty"`
	SdkMessageType     string   `json:"sdk_message_type,omitempty"`
	AllowList          SetModel `json:"allow_list,omitempty"`
	AuthorizationLogic string   `json:"authorization_logic,omitempty"`
	DenyList           SetModel `json:"deny_list,omitempty"`
	Limit              int64    `json:"limit,omitempty"`
	SpecialValue       string   `json:"special_value,omitempty"`
}

type SetModel struct {
	Set []GrantModel `json:"#set"`
}

type ActiveGrantsMapModel struct {
	ActiveGrants [][]ActiveGrantsModel `json:"#map"`
}

type StateModel struct {
	Meta            interface{}          `json:"#meta"`
	ActionTaken     ActionModel          `json:"action_taken"`
	ActiveGrantsMap ActiveGrantsMapModel `json:"active_grants"`
	ExpiredGrants   SetModel             `json:"expired_grants"`
	NumExecs        int                  `json:"num_execs"`
	NumGrants       int                  `json:"num_grants"`
	OutcomeStatus   string               `json:"outcome_status"`
}

type TraceJsonStruct struct {
	Meta   interface{}   `json:"#meta"`
	Vars   []interface{} `json:"vars"`
	States []StateModel  `json:"states"`
}

func getValAddrList(names []string) []sdk.ValAddress {
	addrList := make([]sdk.ValAddress, len(names))
	for i, name := range names {
		addrList[i] = validators[name]
	}

	return addrList
}

//func cosmosUrlToModelMsg(msg_type )
func cosmosAuthorizationFromModel(grantPayload GrantPayloadModel, msgType string, t *testing.T) (authz.Authorization, error) {
	var authorization authz.Authorization
	var err error
	err = nil
	switch grantPayload.AuthorizationLogic {
	case string(Stake):
		{

			allowList := getValAddrList(grantPayload.AllowList.Validators)
			denyList := getValAddrList(grantPayload.DenyList.Validators)

			var stakingAuthorizationType stakingtypes.AuthorizationType
			switch msgType {
			case string(MsgDelegate):
				stakingAuthorizationType = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_DELEGATE
			case string(MsgUndelegate):
				stakingAuthorizationType = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNDELEGATE
			case string(MsgRedelegate):
				stakingAuthorizationType = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_REDELEGATE
			default:
				stakingAuthorizationType = stakingtypes.AuthorizationType_AUTHORIZATION_TYPE_UNSPECIFIED
			}

			var limitAddr *sdk.Coin
			if grantPayload.SpecialValue == "inf" {
				limitAddr = nil
			} else {
				limit := sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(grantPayload.Limit))
				limitAddr = &limit
			}

			authorization, err = stakingtypes.NewStakeAuthorization(
				allowList,
				denyList,
				stakingAuthorizationType,
				limitAddr)

			authorizationMsgTypeURL := modelMsgToCosmosURL[msgType]
			if err == nil {
				require.Equal(t, authorization.MsgTypeURL(), authorizationMsgTypeURL)
			}

		}
	case string(Send):
		{
			limit := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(int64(grantPayload.Limit))))
			authorization = banktypes.NewSendAuthorization(limit)
		}
	case string(Generic):
		{

			msg_url := modelMsgToCosmosURL[msgType]
			authorization = authz.NewGenericAuthorization(msg_url)
		}
	default:
		sdkerrors.ErrInvalidType.Wrapf("modelator: Invalid type for the authorization logic")

	}

	return authorization, err

}

func giveSpecifiedGrant(actionTaken ActionModel, outcome string, t *testing.T) {
	grantPayload := actionTaken.GrantPayload
	authorization, authzErr := cosmosAuthorizationFromModel(grantPayload, actionTaken.Grant.SdkMessageType, t)
	if !(authzErr == nil) {
		require.Equal(t, outcome, GRANT_FAILED)
		return
	}
	if outcome == GRANT_SUCCESS {

		log.WithFields(log.Fields{
			"limit":         grantPayload.Limit,
			"authorization": authorization,
		}).Debug("Upon GRANT_SUCCESS")
		require.NotNil(t, authorization)
		require.NoError(t, authorization.ValidateBasic())
	}

	granterAddr := accountNameToAddress[actionTaken.Grant.Granter]
	granteeAddr := accountNameToAddress[actionTaken.Grant.Grantee]
	now := ctx.BlockHeader().Time
	require.NotNil(t, now)

	// this is a special case check: it is due to the fact that a message
	// with granter == grantee will fail to be processed already in the authz/msgs.go
	// (thus, there will be no check upon calling the SaveGrant function)
	// Perhaps one solution would be to change the level of this driver's implementation
	// as discussed in issue https://github.com/informalsystems/mbt/issues/142
	if actionTaken.Grant.Granter == actionTaken.Grant.Grantee {
		require.Equal(t, outcome, GRANT_FAILED)
		return
	}
	err := app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, authorization, now.Add(time.Hour))
	require.NoError(t, err)

	msgAuthorizationType := modelMsgToCosmosURL[actionTaken.Grant.SdkMessageType]

	log.WithFields(log.Fields{
		"msgAuthorizationType": msgAuthorizationType,
	}).Debug("asking for clean authorization")
	storedAuthz, _ := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, msgAuthorizationType)

	if outcome == GRANT_SUCCESS {
		require.NotNil(t, storedAuthz)
	} else if outcome == GRANT_FAILED {
		require.Nil(t, storedAuthz)
	} else {
		require.True(t, false)
	}

}

func revokeSpecifiedGrant(actionTaken ActionModel, outcome string, t *testing.T) {
	granterAddr := accountNameToAddress[actionTaken.Grant.Granter]
	granteeAddr := accountNameToAddress[actionTaken.Grant.Grantee]
	authorizationType := modelMsgToCosmosURL[actionTaken.Grant.SdkMessageType]

	authorizationOld, _ := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, authorizationType)

	log.WithFields(
		log.Fields{
			"granter":            granterAddr,
			"grantee":            granteeAddr,
			"authorization type": authorizationType,
		}).Debug("revoking authorization")
	err := app.AuthzKeeper.DeleteGrant(ctx, granteeAddr, granterAddr, authorizationType)

	if outcome == REVOKE_FAILED {
		require.Error(t, err)
	} else if outcome == REVOKE_SUCCESS {
		require.NoError(t, err)
	} else {
		require.True(t, false)
	}

	authorization, _ := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, authorizationType)
	if outcome == REVOKE_SUCCESS {
		require.Nil(t, authorization)
	} else if outcome == REVOKE_FAILED {
		require.Equal(t, authorization, authorizationOld)
	} else {
		require.True(t, false)
	}

}

func expireSpecifiedGrant(actionTaken ActionModel, outcome string, t *testing.T) {
	granterAddr := accountNameToAddress[actionTaken.Grant.Granter]
	granteeAddr := accountNameToAddress[actionTaken.Grant.Grantee]
	authorizationType := modelMsgToCosmosURL[actionTaken.Grant.SdkMessageType]
	authorization, _ := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, authorizationType)
	require.NotNil(t, authorization)
	now := ctx.BlockHeader().Time
	require.NotNil(t, now)
	// expiring a grant is done by updating the previous authorization by a new one with an immediate expiration
	err := app.AuthzKeeper.SaveGrant(ctx, granteeAddr, granterAddr, authorization, now)
	require.NoError(t, err)
	updatedTime := now.Add(time.Minute)
	ctx = ctx.WithBlockTime(updatedTime)

}

func executeSpecifiedGrant(actionTaken ActionModel, outcome string, t *testing.T) {
	granteeAddr := accountNameToAddress[actionTaken.Grant.Grantee]
	granterAddr := accountNameToAddress[actionTaken.Grant.Granter]
	grantPayload := actionTaken.GrantPayload
	var message sdk.Msg
	modelMsgType := actionTaken.ExecMessage.Message_Type
	msgTypeURL := modelMsgToCosmosURL[modelMsgType]

	storedAuthz, _ := app.AuthzKeeper.GetCleanAuthorization(ctx, granteeAddr, granterAddr, msgTypeURL)
	if outcome == NONEXISTENT_GRANT_EXEC || outcome == EXPIRED_AUTH_EXEC {
		require.Nil(t, storedAuthz)
		return
	}
	authorization, _ := cosmosAuthorizationFromModel(grantPayload, actionTaken.Grant.SdkMessageType, t)
	require.Equal(t, authorization, storedAuthz)

	switch actionTaken.ExecMessage.Message_Type {

	case string(MsgSend):
		amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(int64(actionTaken.ExecMessage.Amount))))
		defaultRecipient := accountNameToAddress["default_recipient"]
		message = banktypes.NewMsgSend(granteeAddr, defaultRecipient, amount)

	case string(MsgDelegate):
		amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(actionTaken.ExecMessage.Amount))
		valAddress := validators[actionTaken.ExecMessage.Validator]
		message = stakingtypes.NewMsgDelegate(granterAddr, valAddress, amount)

	case string(MsgUndelegate):
		amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(actionTaken.ExecMessage.Amount))
		valAddress := validators[actionTaken.ExecMessage.Validator]
		message = stakingtypes.NewMsgUndelegate(granterAddr, valAddress, amount)

	case string(MsgRedelegate):
		amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, int64(actionTaken.ExecMessage.Amount))
		valAddress := validators[actionTaken.ExecMessage.Validator]
		newValAddress := validators[actionTaken.ExecMessage.NewValidator]
		message = stakingtypes.NewMsgBeginRedelegate(granterAddr, valAddress, newValAddress, amount)

	default:
		sdkerrors.ErrInvalidType.Wrapf("modelator: Invalid message type")
	}

	resp, err := authorization.Accept(ctx, message)
	if outcome == SUCCESSFUL_AUTH_EXEC {
		require.NoError(t, err)
		require.Equal(t, actionTaken.ExecOutcome.Accept, resp.Accept)
		require.Equal(t, actionTaken.ExecOutcome.Delete, resp.Delete)
		updatedGrant, _ := cosmosAuthorizationFromModel(actionTaken.ExecOutcome.Updated, actionTaken.Grant.SdkMessageType, t)
		if grantPayload.AuthorizationLogic == string(Generic) {
			require.Nil(t, resp.Updated)
		} else {
			require.Equal(t, updatedGrant, resp.Updated)
		}
	} else if outcome == GRANT_SPENT {
		require.NoError(t, err)
		require.Equal(t, actionTaken.ExecOutcome.Accept, resp.Accept)
		require.Equal(t, actionTaken.ExecOutcome.Delete, resp.Delete)
		require.Nil(t, resp.Updated)
	} else {
		require.Error(t, err)
	}

}

func executeAppropriateAction(state StateModel, t *testing.T) {
	actionTaken := state.ActionTaken
	outcome := state.OutcomeStatus
	switch ActionType(actionTaken.ActionType) {
	case GiveGrant:
		giveSpecifiedGrant(actionTaken, outcome, t)
	case RevokeGrant:
		revokeSpecifiedGrant(actionTaken, outcome, t)
	case ExpireGrant:
		expireSpecifiedGrant(actionTaken, outcome, t)
	case ExecuteGrant:
		executeSpecifiedGrant(actionTaken, outcome, t)
	case "":
		// nothing here, "" is init state
	default:
		log.WithFields(log.Fields{
			"action type": actionTaken.ActionType,
		}).Fatal("no appropriate action found")

	}

}

func testEnvironmentSetup(t *testing.T) {
	app = simapp.Setup(false)
	ctx = app.BaseApp.NewContext(false, tmproto.Header{})
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	PKs = simapp.CreateTestPubKeys(5)

	// connects to the accounts' names from the model
	accountNameToAddress = make(map[string]sdk.AccAddress)
	accountsAddresses := simapp.AddTestAddrs(app, ctx, 4, sdk.NewInt(1000))
	accountNameToAddress["A"] = accountsAddresses[0]
	accountNameToAddress["B"] = accountsAddresses[1]
	accountNameToAddress["C"] = accountsAddresses[2]
	accountNameToAddress["default_recipient"] = accountsAddresses[3]

	// c onnects to the validators' names from the model
	validators = make(map[string]sdk.ValAddress)
	validatorAddresses := simapp.ConvertAddrsToValAddrs(simapp.AddTestAddrs(app, ctx, 3, sdk.NewInt(1000)))
	validators["X"] = validatorAddresses[0]
	validators["Y"] = validatorAddresses[1]
	validators["Z"] = validatorAddresses[2]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	for i, valAddr := range validatorAddresses {
		tstaking.CreateValidator(valAddr, PKs[i], sdk.NewInt(10), true)
	}

	// make so that each account delegates with each of the validators.
	// We do so because the model only cares about whether a certain message has an authorization
	// (for instance, we want to be able to undelegate from state 0)
	for _, accAddr := range accountsAddresses {
		for _, valAddr := range validatorAddresses {
			tstaking.Delegate(accAddr, valAddr, sdk.NewInt(50))
		}
	}

	modelMsgToCosmosURL = map[string]string{

		string(MsgDelegate):   sdk.MsgTypeURL(&stakingtypes.MsgDelegate{}),
		string(MsgUndelegate): sdk.MsgTypeURL(&stakingtypes.MsgUndelegate{}),
		string(MsgRedelegate): sdk.MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}),
		string(MsgSend):       sdk.MsgTypeURL(&banktypes.MsgSend{}),
		string(MsgOther):      sdk.MsgTypeURL(&govtypes.MsgVote{}),
	}

	cosmosURLToModelMsg = make(map[string]string)

	for k, v := range modelMsgToCosmosURL {
		cosmosURLToModelMsg[v] = k
	}

}

/*
 - figure out the best way for different levels of logging in Go
 - generate a set of tests for all possible behaviors from the model

*/
func TestExecuteItfJson(t *testing.T) {

	// Will log anything that is info or above (warn, error, fatal, panic). Default.
	// log.SetLevel(log.DebugLevel)

	umbrellaDir := "traces"
	// umbrellaDir := "purgedTraces"
	testDirs, err := os.ReadDir(umbrellaDir)
	if err != nil {

		log.Fatal(err)
	}
	// iterate over all directories corresponding to different behavior scenarios
	for _, testDir := range testDirs {
		currentDir := filepath.Join(umbrellaDir, testDir.Name())
		if testDir.IsDir() {
			testFiles, err := os.ReadDir(currentDir)
			if err != nil {
				log.Fatal(err)
			}
			// iterate over all trace-files
			for _, testFile := range testFiles {
				filePath := filepath.Join(currentDir, testFile.Name())
				log.WithFields(
					log.Fields{
						"file path": filePath,
					}).Info("Currently testing.")
				traceFile, err := os.Open(filePath)
				if err != nil {
					log.Error(err)
				}
				defer traceFile.Close()

				byteValue, _ := ioutil.ReadAll(traceFile)
				var jsonData TraceJsonStruct
				json.Unmarshal(byteValue, &jsonData)

				// setup the testing environment
				testEnvironmentSetup(t)

				for _, state := range jsonData.States {
					executeAppropriateAction(state, t)
				}

			}

		}

	}

}
