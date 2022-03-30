---- MODULE statusCodes ----
EXTENDS Integers

CONSTANT 
    \* @type: Int;
    MaxPossibleAmount

INFINITY == "inf"
INAPPROPRIATE_AUTH == "inappropriate_auth"
INAPPROPRIATE_AUTH_GENERIC == "inappropriate_auth_generic"
INAPPROPRIATE_AUTH_SEND == "inappropriate_auth_send"

INAPPROPRIATE_AUTH_STAKE_NOT_ALLOW == "inappropriate_auth_stake_not_allow"
INAPPROPRIATE_AUTH_STAKE_DENY == "inappropriate_auth_stake_deny"
INAPPROPRIATE_AUTH_FOR_MESSAGE == "message_not_supported_by_the_authorization"


SUCCESSFUL_AUTH_EXEC == "successful_auth_exec"
INSUFFICIENT_GRANT_EXEC == "insufficient_auth_exec"
TRIVIAL_AUTH == "trivial_auth"
NONEXISTENT_GRANT_EXEC == "non_existent_auth"
EXPIRED_GRANT == "expired_grant"
EXPIRED_AUTH_EXEC == "tried to execute an expired grant"
GRANT_FAILED == "grant_failed"
REVOKE_FAILED == "revoke_failed"
GRANT_SUCCESS == "grant_success"
REVOKE_SUCCESS == "revoke_success"
TEST == "TEST"

SPENT_AUTH_EXEC == "grant_spent"
NO_OP == "/"
APPROPRIATE == "ok"
RESERVED == "reserved"

EMPTY_INT_VALUE == -1




\* AUTHORIZATION LOGIC TYPES
SEND == "send"
STAKE == "stake"
GENERIC == "generic"

\* SDK MESSAGE TYPES
MSG_SEND == "msg_send"
MSG_DELEGATE == "msg_delegate"
MSG_UNDELEGATE == "msg_undelegate"
MSG_REDELEGATE == "msg_redelegate"
MSG_ALPHA == "msg_alpha"
STAKING_MESSAGES == {MSG_DELEGATE, MSG_UNDELEGATE, MSG_REDELEGATE}

\* @type: (Str) => Set(Str);
SUPPORTED_AUTHORIZATIONS(msg_type) ==
        CASE msg_type = MSG_SEND -> {SEND, GENERIC}
          [] msg_type = MSG_DELEGATE -> {STAKE, GENERIC}
          [] msg_type = MSG_UNDELEGATE -> {STAKE, GENERIC}
          [] msg_type = MSG_REDELEGATE -> {STAKE, GENERIC}
          [] OTHER -> {GENERIC}    


====