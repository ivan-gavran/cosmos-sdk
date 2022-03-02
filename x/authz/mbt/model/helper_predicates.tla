---- MODULE helper_predicates ----

EXTENDS 
    generic_authz,
    stake_authz,
    authz_typedefs, 
    send_authz




(*
    Getting the payload associated with grant `g`.    
*)
\* @type: (GRANT, GRANT -> GRANT_PAYLOAD) => GRANT_PAYLOAD;
_GetGrantPayload(grant, active_grants) ==
    IF grant \in DOMAIN active_grants
    THEN 
        active_grants[grant]
    ELSE 
        EmptyPayload


\* @type: (GRANT_PAYLOAD) => Bool;
_GrantPayloadSaneValues(grant_payload) == 
    IF grant_payload.authorization_logic /= STAKE
    THEN
        grant_payload.special_value = ""
    ELSE 
        \/ (grant_payload.special_value /= INFINITY /\ grant_payload.limit > 0)
        \/ (grant_payload.special_value = INFINITY /\ grant_payload.limit = -1)    



(*
This is sa predicate-split for different authorization types (send, stake, generic)
*)
\* @type: (GRANT_PAYLOAD, MESSAGE) => Bool;
_IsGrantSpent(g_payload, msg) ==            
    CASE  g_payload.authorization_logic = SEND -> _IsGrantSpentSend(g_payload, msg)    
      [] g_payload.authorization_logic = STAKE -> _IsGrantSpentStake(g_payload, msg)
      \* working only with generic, send, stake; other branch should never be reached
      [] OTHER -> _IsGrantSpentGeneric(g_payload, msg)             
    
    


(*
This is sa predicate-split for different authorization types (send, stake, generic)
*)
\* @type: (GRANT_PAYLOAD, MESSAGE) => Str;
_IsGrantAppropriate(g_payload, msg) ==    
    IF g_payload.authorization_logic \notin SUPPORTED_AUTHORIZATIONS(msg.message_type) THEN INAPPROPRIATE_AUTH_FOR_MESSAGE
    ELSE IF g_payload.authorization_logic = SEND THEN _IsGrantAppropriateSend(g_payload, msg)
    ELSE IF g_payload.authorization_logic = STAKE THEN _IsGrantAppropriateStake(g_payload, msg)
    ELSE  _IsGrantAppropriateGeneric(g_payload, msg)

(*
This is sa predicate-split for different authorization types (send, stake, generic)
*)
\* @type: (GRANT_PAYLOAD, MESSAGE) => GRANT_PAYLOAD;
_UpdateAuth(g_payload, msg) ==     
    CASE g_payload.authorization_logic = STAKE -> _UpdateAuthStake(g_payload, msg)
      [] g_payload.authorization_logic = SEND -> _UpdateAuthSend(g_payload, msg)            
      [] OTHER -> _UpdateAuthGeneric(g_payload, msg)
        



\* @type: (Str, Str, MESSAGE) => GRANT;
_GetGrant(sender, signer, msg) ==
    [granter |-> signer, grantee |-> sender, sdk_message_type |-> msg.message_type]




====