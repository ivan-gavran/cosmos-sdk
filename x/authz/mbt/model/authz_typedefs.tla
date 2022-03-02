---- MODULE authz_typedefs ----
EXTENDS 
    Integers,
    (* I want to keep status codes around as a way to later produce tests referring to executions succeeding, passing, 
     grants timeouting etc.*)
    statusCodes

CONSTANT
    \* @type: Set(Str);  
    Identifiers,    
    \* @type: Set(Str);
    Validators,
    \* @type: Set(Str);
    AllSDKMessageTypes,
    \* @type: Set(Str);
    AllAuthorizationLogics
    

\* IMPORTANT: pay attention to spaces before the "@typeAlias". These are important
\* as a quick fix to an existing bug in Apalache: https://github.com/informalsystems/apalache/issues/1304

(* 
 @typeAlias: GRANT = [ 
    granter: Str, 
    grantee: Str, 
    sdk_message_type: Str 
    ];

 @typeAlias: GRANT_PAYLOAD = [ 
    limit: Int,     
    allow_list: Set(Str),
    deny_list: Set(Str),
    special_value: Str,
    authorization_logic: Str
    ];

 @typeAlias: MESSAGE = [
    message_type: Str,    
    validator: Str,
    new_validator: Str,
    amount: Int
];

 @typeAlias: ACTION = [
    grant: GRANT,    
    grant_payload: GRANT_PAYLOAD,    
    exec_outcome: EXEC_OUTCOME,    
    action_type: Str,
    exec_message: MESSAGE
];

 @typeAlias: STATE = [     
  active_grants: GRANT -> GRANT_PAYLOAD,  
  expired_grants: Set(GRANT),
  num_grants: Int,
  num_execs: Int,
  outcome_status: Str,
  action_taken: ACTION
];

 @typeAlias: EXEC_OUTCOME = [
     accept: Bool,
     delete: Bool,
     updated: GRANT_PAYLOAD,
     description: Str
 ];

*)

  EmptyMessage == [
    message_type |-> "",    
    validator |-> "",
    new_validator |-> "",
    amount |-> -1
]

EmptyGrant == [granter |-> "", grantee |-> "", sdk_message_type |-> ""]

\* @type: Set(Str);
EmptySetOfStrings == {}




\* @type: GRANT_PAYLOAD;
EmptyPayload == [
    limit |-> 0,     
    allow_list |-> EmptySetOfStrings,
    deny_list |-> EmptySetOfStrings,
    special_value |-> "",
    authorization_logic |-> ""
]

EmptyExecOutcome == [
    accept |-> FALSE,
    delete |-> FALSE,
    updated |-> EmptyPayload,
    description |-> ""
]


EmptyAction == 
    [
        action_type |-> "",
        grant |-> EmptyGrant,        
        grant_payload |-> EmptyPayload,
        exec_outcome |-> EmptyExecOutcome,
        exec_message |-> EmptyMessage
    ]


AllGrants == [
    granter: Identifiers,
    grantee: Identifiers,
    sdk_message_type: AllSDKMessageTypes
]

Messages == 
    [
    message_type: AllSDKMessageTypes,    
    \* this field is used only with stake authorizations
    validator: Validators \union {""},
    new_validator: Validators \union {""},
    amount: 0..MaxPossibleAmount \union {EMPTY_INT_VALUE}
]

GrantPayloads == [
    limit: -1..MaxPossibleAmount,     
    allow_list: SUBSET Validators,
    deny_list: SUBSET Validators,    
    special_value: {INFINITY} \union {""},
    authorization_logic: AllAuthorizationLogics
]


\* @type: (GRANT, GRANT_PAYLOAD) => Bool;
_GrantPayloadTypes(g, g_payload) ==    
    
    IF g_payload.authorization_logic = GENERIC
    THEN  
        /\ g_payload.limit = -1
        /\ g_payload.allow_list = EmptySetOfStrings
        /\ g_payload.deny_list = EmptySetOfStrings
        /\ g_payload.special_value = ""          
    ELSE IF g_payload.authorization_logic = SEND
    THEN         
        /\ g_payload.allow_list = EmptySetOfStrings
        /\ g_payload.deny_list = EmptySetOfStrings
        /\ g_payload.special_value = ""                
    ELSE \* , that is, IF g_payload.authorization_logic = STAKE
        \/ (g_payload.allow_list = EmptySetOfStrings /\ g_payload.deny_list /= EmptySetOfStrings)
        \/ (g_payload.allow_list /= EmptySetOfStrings /\ g_payload.deny_list = EmptySetOfStrings)

    

\* @type: (MESSAGE) => Bool;
_ExecMessagePayloadTypes(msg) ==    
    \/
        /\ msg.message_type = MSG_SEND    
        /\ msg.validator = ""  
        /\ msg.new_validator = ""              
        /\ msg.amount > 0
    \/
        /\ msg.message_type \in {MSG_DELEGATE, MSG_UNDELEGATE}
        /\ msg.validator /= ""
        /\ msg.new_validator = ""
        /\ msg.amount > 0

    \/
        /\ msg.message_type = MSG_REDELEGATE
        /\ msg.validator /= ""
        /\ msg.new_validator /= ""
        /\ msg.amount > 0
    
        




====