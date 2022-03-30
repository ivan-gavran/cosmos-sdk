---- MODULE generic_authz ----
EXTENDS statusCodes

\* generic grant can never be spent
_IsGrantSpentGeneric(g_payload, msg) == FALSE

\* generic grant is always appropriate
\* @type: (GRANT_PAYLOAD, MESSAGE) => Str;
_IsGrantAppropriateGeneric(g, msg) == APPROPRIATE

(*
Generic grant always remains the same and is never updated.
*)
\* @type: (GRANT_PAYLOAD, MESSAGE) => GRANT_PAYLOAD;
_UpdateAuthGeneric(g_payload, msg) == g_payload

  


====