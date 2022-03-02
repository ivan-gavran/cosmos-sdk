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


====