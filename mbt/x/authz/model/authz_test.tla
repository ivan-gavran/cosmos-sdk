---- MODULE authz_test ----
EXTENDS authz
    



Init ==
    /\ active_grants = [g \in {} |-> EmptyPayload]
    /\ expired_grants = {}
    /\ num_grants = 0
    /\ num_execs = 0
    /\ outcome_status = ""
    /\ action_taken = EmptyAction


Next == 
    \/ \E g \in AllGrants:                
            \/ RevokeGrant(g)
            \/ ExpireGrant(g)
            \/ \E grant_payload \in GrantPayloads: GiveGrant(g, grant_payload)            
    \/ \E sender, signer \in Identifiers, msg \in Messages: 
            Execute(sender, signer, msg)
            

\* view to be used with TLC, to avoid enumerating same states
View == <<active_grants, expired_grants, outcome_status>>

\* view to be used with Apalache, to create a diverse set of counterexamples
CounterexamplesView == active_grants



(*
======== TESTS =======
*)

(*
ideas for more tests:
 - there is a grant that contains limit equals to zero
 - there is a send/stake grant that contains limit equal to -1
*)

(* <--- *)
GrantSuccess ==    
   /\ outcome_status = GRANT_SUCCESS
   /\ num_grants > 1

GrantSuccessCEX == ~ GrantSuccess
(* ---> *)


(* <--- *)
GrantFailed ==
    /\ num_grants > 3
    /\ outcome_status = GRANT_FAILED

GrantFailedCEX == ~ GrantFailed
(* ---> *)



Test == 
    outcome_status /= TEST

(* <--- *)
\* @type: Seq(STATE) => Bool;
GrantFailedFollowedBySuccess(trace) ==    
    \E i \in 1..5:
        /\ 
            LET state1 == trace[i] IN 
            LET state2 == trace[i+1] IN
            /\ state1.outcome_status = GRANT_FAILED
            /\ state2.outcome_status = GRANT_SUCCESS
            /\ Len(trace) >= i+1


\* @type: Seq(STATE) => Bool;
GrantFailedFollowedBySuccessCEX(trace) == ~ GrantFailedFollowedBySuccess(trace)
(* ---> *)




(* <--- *)
\* @type: Seq(STATE) => Bool;
ExpiredGrantGivenAgainExec(trace) ==    
    \E i, j, k \in 1..5:
        /\ 
            LET state1 == trace[i] IN 
            LET state2 == trace[j] IN
            LET state3 == trace[k] IN
            LET grant1 == state1.action_taken.grant IN
            LET grant2 == state2.action_taken.grant IN
            LET grant3 == state3.action_taken.grant IN            
            /\ i < j
            /\ j < k
            /\ state1.outcome_status = EXPIRED_GRANT            
            /\ state2.outcome_status = GRANT_SUCCESS
            /\ state3.outcome_status = SUCCESSFUL_AUTH_EXEC
            /\ grant1 = grant2
            /\ grant2 = grant3
            /\ Len(trace) >= k


\* @type: Seq(STATE) => Bool;
ExpiredGrantGivenAgainExecCEX(trace) == ~ ExpiredGrantGivenAgainExec(trace)
(* ---> *)

\* @type: Seq(STATE) => Bool;
ExpiredAuthV2(trace) ==
    \E i \in 1..5:
    /\
        LET state1 == trace[i] IN 
        state1.outcome_status = EXPIRED_GRANT
    /\ Len(trace) > i

ExpiredAuthV2CEX(trace) == ~ ExpiredAuthV2(trace)


(* <--- *)
\* @type: Bool;
ExpiredAuth ==
    outcome_status = EXPIRED_GRANT

ExpiredAuthCEX == ~ ExpiredAuth
(* ---> *)


(* <--- *)
\* @type: Bool;
ExecOfExpiredAuth ==
    outcome_status = EXPIRED_AUTH_EXEC

ExecOfExpiredAuthCEX == ~ ExecOfExpiredAuth
(* ---> *)


(* <--- *)
\* @type: Bool;
SuccessfulExec ==
    outcome_status = SUCCESSFUL_AUTH_EXEC

SuccessfulExecCEX == ~ SuccessfulExec
(* ---> *)

(* <--- *)
InfValueExecution ==
    /\ outcome_status = SUCCESSFUL_AUTH_EXEC
    /\ action_taken.grant_payload.special_value = INFINITY

InfValueExecutionCEX == ~ InfValueExecution
(* ---> *)



(* <--- *)
\* @type: Bool;
TrivialAuth ==
    outcome_status = TRIVIAL_AUTH

TrivialAuthCEX == ~ TrivialAuth
(* ---> *)



(* <--- *)
INAPPROPRIATE_AUTHS == {INAPPROPRIATE_AUTH_GENERIC, INAPPROPRIATE_AUTH_STAKE_DENY, INAPPROPRIATE_AUTH_STAKE_NOT_ALLOW, INAPPROPRIATE_AUTH_SEND, INAPPROPRIATE_AUTH}

\* @type: Bool;
InappropriateAuth ==
    outcome_status \in INAPPROPRIATE_AUTHS

InappropriateAuthCEX == ~ InappropriateAuth
(* ---> *)

(* <--- *)
\* @type: Bool;
InappropriateAuthStakeDeny ==
    outcome_status = INAPPROPRIATE_AUTH_STAKE_DENY

InappropriateAuthStakeDenyCEX == ~ InappropriateAuthStakeDeny
(* ---> *)


(* <--- *)
\* @type: Bool;
RevokeSuccess ==
    /\ outcome_status = REVOKE_SUCCESS
    /\ Cardinality(DOMAIN active_grants) > 2

RevokeSuccessCEX == ~ RevokeSuccess
(* ---> *)


(* <--- *)
\* @type: Bool;
RevokeFailure ==
    /\ outcome_status = REVOKE_FAILED
    /\  Cardinality(DOMAIN active_grants) > 2

RevokeFailureCEX == ~ RevokeFailure
(* ---> *)



(* <--- *)
\* @type: Bool;
ExecGrantSpent ==
    outcome_status = SPENT_AUTH_EXEC    

ExecGrantSpentCEX == ~ ExecGrantSpent
(* ---> *)



(* <--- *)
\* @type: Bool;
ExecGrantInsufficient ==
    outcome_status = INSUFFICIENT_GRANT_EXEC

ExecGrantInsufficientCEX == ~ ExecGrantInsufficient
(* ---> *)



(* <--- *)
\* @type: Bool;
NonexistentAuth ==
    outcome_status = NONEXISTENT_GRANT_EXEC


NonexistentAuthCEX == ~ NonexistentAuth
(* ---> *)

====