-------------------------- MODULE authz ----------------------------
(*
 
This is a model for the authz module, based on the natural language specification: https://github.com/cosmos/cosmos-sdk/tree/75f5dd7da457d2ba1f3c2f01c7c9bf208e52be36/x/authz/spec
It models giving grants, revoking grants and using grants in messages.
The logic of particular kinds of grant authorizations (Send, Stake, Generic) is described in separate authorization files (send_authz.tla, stake_authz.tla).
The interface between this file and the authorization files is given in helper_predicates.tla

 TODO: 
         

     - a weird bug with @ notations (send_authz.tla, function UpdateAuthSend, line 10, ). Apalache v.0.20.2: does not typecheck, even though
     it should be equivalaent to the explicit constructor.

     - perhaps it would be meaningful to explicitly model all failures by a special failure state (instead of forbidding those situation in the
     model). That way, more interesting behavior could be tested
 
 *) 

EXTENDS 
    Integers, 
    FiniteSets, 
    Sequences,            
    helper_predicates,    
    TLC


\* state variables
VARIABLE
  (* a set containing all active grants *)
  \* @type: GRANT -> GRANT_PAYLOAD;  
  active_grants,
  (* a set of grants that timed out. Modelling this explicitly so that the functionality of removing expired grants can be tested *)
  \* @type: Set(GRANT);
  expired_grants
  

\*   statistical and debugging variables
VARIABLE 
    \* @type: Int; 
    num_grants,
    \* @type: Int;
    num_execs,
    \* @type: Str;
    outcome_status,  
    \* @type: ACTION;
    action_taken
  


\* @type: (GRANT, GRANT_PAYLOAD) => Bool;
_GiveGrantPrecondition(g, g_payload) ==            
    \* must be a supported authorization
    /\ g_payload.authorization_logic \in SUPPORTED_AUTHORIZATIONS(g.sdk_message_type)
    \* granter and grantee must be distinct
    /\ g.granter /= g.grantee
    \* exactly one of (allow_list, deny_list) has to be non-empty
    /\ (g_payload.authorization_logic = STAKE) => 
        \/ (g_payload.allow_list = EmptySetOfStrings /\ g_payload.deny_list /= EmptySetOfStrings)
        \/ (g_payload.allow_list /= EmptySetOfStrings /\ g_payload.deny_list = EmptySetOfStrings) 

        
(*
Create a new grant `g` with the payload `g_payload`. This is modelled by updating the function `active_grants`.
Giving a grant fails if granter and grantee have the same address 
(as specified at https://github.com/cosmos/cosmos-sdk/blob/75f5dd7da4/x/authz/spec/03_messages.md)
*)
\* @type: (GRANT, GRANT_PAYLOAD) => Bool;
GiveGrant(g, g_payload) ==    
    \* case-split: whether the grant succeeds or fails
    /\  
        \/  \* first option: grant failure
            /\ ~ _GiveGrantPrecondition(g, g_payload)
            /\ outcome_status' = GRANT_FAILED        
            /\ UNCHANGED <<active_grants, expired_grants>>             
        \/
            /\ _GiveGrantPrecondition(g, g_payload)
            \* Because g_payload has type which is a supertype for all send, stake, and generic authorization, 
            \* this predicate forces unused fields of g_payload to be set to default values
            /\ _GrantPayloadTypes(g, g_payload)
            /\ _GrantPayloadSaneValues(g_payload)
            \* update of state variables
            /\ active_grants' = [k \in DOMAIN active_grants \union {g} |-> 
                IF k = g THEN g_payload ELSE active_grants[k]]     
            /\ outcome_status' = GRANT_SUCCESS
            \* if there was a previously expired grant, remove it now (the fresh grant overwrites this previously expired one)
            /\ expired_grants' = expired_grants \ {g}    

    \* bookkeeping
    /\ action_taken' = [
        action_type |-> "give grant",        
        grant |-> g,        
        grant_payload |-> g_payload,
        exec_outcome |-> EmptyExecOutcome,
        \* exec_message field is here only for the Exec message. Thus, we are here setting it to a default value
        exec_message |-> EmptyMessage
        ]
    /\ num_grants' = num_grants + 1
    /\ UNCHANGED num_execs


\* @type: (GRANT) => Bool;
_RevokeGrantPrecondition(g) ==
    /\ g.granter /= g.grantee
    /\ g \in DOMAIN active_grants
    /\ g.sdk_message_type /= ""

(*
Revoking an existing grant `g`.
The predicate returns False if there is no grant `g` or if granter's and grantee's address is the same.
*)
\* @type: (GRANT) => Bool;
RevokeGrant(g) ==    
    /\ 
        \* case-split on precondition       
        \/  \* if precondition is NOT satisfied 
            /\ _RevokeGrantPrecondition(g)
            /\ outcome_status' = REVOKE_FAILED
            /\ UNCHANGED <<expired_grants, active_grants>>
        \/ \* if precondition is satisfied
            /\ _RevokeGrantPrecondition(g)
            \* active grants is the same, but `g` is removed from its domain        
            /\ active_grants' = [k \in DOMAIN active_grants \ {g} |-> active_grants[k]]
            \* the grant `g` is removed from the set of expired grants (in case it was there)
            /\ expired_grants' = expired_grants \ {g}
            /\ outcome_status' = REVOKE_SUCCESS            

    \* bookkeeping
    /\ action_taken' = [
        action_type |-> "revoke grant",
        grant |-> g,
        \* both grant_payload and exec_message are irrelevant fields for the RevokeGrant ---> thus, they are set to their defaults
        grant_payload |-> EmptyPayload,        
        exec_message |-> EmptyMessage
        ]
    /\ UNCHANGED <<num_execs, num_grants>>

    


(*
Timeout being reached is abstracted away by the action ExpireGrant. What is lost then is the relation between different grants
and their expirations (in real life, there could be a dependency: if A expires, then definitely B and C have to expire).
This action is a pure environment action and thus it does not remove the grant from a set of active grants (the system will do so
once it notices that the grant expired, upon running the Execute function)
*)
ExpireGrant(g) ==    
    /\ g \in DOMAIN active_grants        
    /\ action_taken' = [
        action_type |-> "expire grant",
        grant |-> g,
        grant_payload |-> EmptyPayload,
        exec_outcome |-> EmptyExecOutcome,
        exec_message |-> EmptyMessage
        ]
    /\ outcome_status' = EXPIRED_GRANT
    /\ expired_grants' = expired_grants \union {g}        
    /\ UNCHANGED <<num_grants, active_grants, num_execs>>




(*
======
The following functions model executing a message, that is: checking if active_grants allow for successful execution.
======
*)


\* @type: (GRANT, GRANT_PAYLOAD, MESSAGE) => EXEC_OUTCOME;
Accept(grant, grant_payload, msg) ==

        LET updated_payload ==  _UpdateAuth(grant_payload, msg) IN         
        LET grant_appropriate == _IsGrantAppropriate(grant_payload, msg) IN
        IF grant \notin DOMAIN active_grants
        THEN [accept |-> FALSE, delete |-> FALSE, updated |-> EmptyPayload, description |-> NONEXISTENT_GRANT_EXEC]
        ELSE IF grant \in expired_grants
        THEN [accept |-> FALSE, delete |-> TRUE, updated |-> EmptyPayload, description |-> EXPIRED_AUTH_EXEC ]
        ELSE IF grant_appropriate /= APPROPRIATE
        THEN [accept |-> FALSE, delete |-> FALSE, updated |-> EmptyPayload, description |-> grant_appropriate]
        ELSE IF (updated_payload.limit = 0 /\ updated_payload.special_value /= INFINITY)
        THEN [accept |-> TRUE, delete |-> TRUE, updated |-> updated_payload, description |-> SPENT_AUTH_EXEC]
        ELSE [accept |-> TRUE, delete |-> FALSE, updated |-> updated_payload, description |-> SUCCESSFUL_AUTH_EXEC]



(*
The predicate which models the execution of a message.
It checks whether there is an appropriate grant for the execution message, but does not model the actual execution and whether or not
the execution itself fails.
*)
\* @type: (Str, Str, MESSAGE) => Bool;
Execute(sender, signer, msg) == 
    \* @type: GRANT;
    LET grant == _GetGrant(sender, signer, msg) IN
    LET grant_payload == _GetGrantPayload(grant, active_grants) IN 
    LET accept_result == Accept(grant, grant_payload, msg) IN
    
    \* makes sure that irrelevant fields of the `msg` variable are set to their default values 
    /\ _ExecMessagePayloadTypes(msg)           
    /\ num_execs' = num_execs + 1        
    /\ UNCHANGED <<num_grants>>    
    /\ action_taken' = [
        action_type |-> "execute grant",
        grant |-> grant,
        grant_payload |-> grant_payload,
        exec_outcome |-> accept_result,
        exec_message |-> msg
        ]
    /\ outcome_status' = accept_result.description
    /\ active_grants' = 
        IF accept_result.delete
        THEN [k \in DOMAIN active_grants \ {grant} |-> active_grants[k]]
        ELSE IF accept_result.accept 
        THEN [active_grants EXCEPT ![grant] = accept_result.updated]            
        ELSE active_grants
    /\ expired_grants' =
        IF accept_result.delete 
        THEN expired_grants \ {grant}
        ELSE expired_grants
        

=============================================================================