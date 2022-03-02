---- MODULE stake_authz ----

EXTENDS 
    statusCodes, 
    Integers


\* @type: (GRANT_PAYLOAD, MESSAGE) => GRANT_PAYLOAD;
_UpdateAuthStake(g_payload, msg) == 
    
    \* per specification, if the payload limit is LEFT EMPTY, then it behaves as infinite limit
    IF g_payload.special_value = INFINITY
    THEN g_payload
    ELSE
    \* IG: WEIRD: why doesn't this expression work?
    \* [g_payload EXCEPT !.limit = (@ - msg.amount)]
    [
        limit |-> g_payload.limit - msg.amount,     
        allow_list |-> g_payload.allow_list,
        deny_list |-> g_payload.deny_list,
        special_value |-> g_payload.special_value,
        authorization_logic |-> g_payload.authorization_logic
    ]

\* @type: (GRANT_PAYLOAD, MESSAGE) => Bool;
_IsGrantSpentStake(g_payload, msg) == 
    /\ g_payload.special_value /= INFINITY
    /\ g_payload.limit = msg.amount            



\* @type: (GRANT_PAYLOAD, MESSAGE) => Str;
_IsGrantAppropriateStake(g_payload, msg) == 
    LET validator_to_check ==
        IF msg.message_type \in {MSG_DELEGATE, MSG_UNDELEGATE}
        THEN msg.validator
        ELSE msg.new_validator
    IN 
    IF g_payload.limit < msg.amount /\ g_payload.special_value /= INFINITY
    THEN INSUFFICIENT_GRANT_EXEC
    ELSE IF validator_to_check \in g_payload.deny_list
    THEN INAPPROPRIATE_AUTH_STAKE_DENY
    ELSE IF validator_to_check \notin g_payload.allow_list
    THEN INAPPROPRIATE_AUTH_STAKE_NOT_ALLOW    
    ELSE APPROPRIATE
    
    

====