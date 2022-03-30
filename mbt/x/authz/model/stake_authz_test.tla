---- MODULE stake_authz_test ----
EXTENDS authz_test


(* <--- *)
\* @type: Bool;
InappropriateAuthValidatorNotInallowList ==
    outcome_status = INAPPROPRIATE_AUTH_STAKE_NOT_ALLOW

InappropriateAuthValidatorNotInallowListCEX == ~ InappropriateAuthValidatorNotInallowList
(* ---> *)



(* <--- *)
\* @type: Bool;
InappropriateAuthValidatorIndenyList ==
    outcome_status = INAPPROPRIATE_AUTH_STAKE_DENY

InappropriateAuthValidatorIndenyListCEX == ~ InappropriateAuthValidatorIndenyList
(* ---> *)


(* <--- *)
\* @type: Bool;
InappropriateAuthWrongStakingAction ==
    outcome_status = INAPPROPRIATE_AUTH_FOR_MESSAGE

InappropriateAuthWrongStakingActionCEX == ~ InappropriateAuthWrongStakingAction
(* ---> *)

SuccessfulExecWithDeny ==
    /\ outcome_status = SUCCESSFUL_AUTH_EXEC
    /\ num_grants > 2
    /\ Cardinality(action_taken.grant_payload.deny_list) > 0

SuccessfulExecWithDenyCEX == ~ SuccessfulExecWithDeny

====