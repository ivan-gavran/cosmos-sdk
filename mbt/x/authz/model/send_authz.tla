---- MODULE send_authz ----
EXTENDS 
    statusCodes, 
    Integers

(*
Updates the available limit for the grant
*)
\* @type: (GRANT_PAYLOAD, MESSAGE) => GRANT_PAYLOAD;
_UpdateAuthSend(g_payload, msg) ==         
    \* IG: WEIRD: why doesn't this expression work?
    \* [g_payload EXCEPT !.limit = (@ - msg.amount)]

    \* @type: GRANT_PAYLOAD;
    LET updated_payload ==
    [
        limit |-> g_payload.limit - msg.amount,     
        allow_list |-> g_payload.allow_list,
        deny_list |-> g_payload.deny_list,        
        special_value |-> g_payload.special_value,
        authorization_logic |-> g_payload.authorization_logic
    ]
    IN updated_payload

\* @type: (GRANT_PAYLOAD, MESSAGE) => Bool;
_IsGrantSpentSend(g_payload, msg) == 
    g_payload.limit = msg.amount            


    
\* @type: (GRANT_PAYLOAD, MESSAGE) => Str;
_IsGrantAppropriateSend(g_payload, msg) == 
    IF g_payload.limit >= msg.amount
    THEN 
        APPROPRIATE
    ELSE
        INSUFFICIENT_GRANT_EXEC
            


====