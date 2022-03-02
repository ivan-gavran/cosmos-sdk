# Authz module specification
This is a TLA+ specification of the authz module, following the module's [natural language spec](https://github.com/cosmos/cosmos-sdk/tree/75f5dd7da457d2ba1f3c2f01c7c9bf208e52be36/x/authz/spec).

## Description of authz
`authz` is an implementation of a Cosmos SDK module, which allows a `granter` account to allow a `grantee` account executing messages of type `msg`.
Different kinds of authorization logics can be defined by implementing the [`Authorization`](https://github.com/cosmos/cosmos-sdk/blob/75f5dd7da4/x/authz/authorizations.go) interface provided by `authz`.
(Because of that, the implementation relevant for `authz` can be found in the [x/authz repo](https://github.com/cosmos/cosmos-sdk/blob/75f5dd7da4/x/authz/keeper/keeper.go), but also in the [x/bank repo](https://github.com/cosmos/cosmos-sdk/blob/75f5dd7da4/x/bank/types/send_authorization.go) and in the [x/staking repo](https://github.com/cosmos/cosmos-sdk/blob/75f5dd7da4/x/staking/types/authz.go)).

### Main concepts 
The `authz`'s main concepts are best described by three kinds of messages it allows:
 - [MsgGrant](https://github.com/cosmos/cosmos-sdk/blob/75f5dd7da457d2ba1f3c2f01c7c9bf208e52be36/x/authz/spec/03_messages.md#MsgGrant) for giving a grant
 - [MsgRevoke](https://github.com/cosmos/cosmos-sdk/blob/75f5dd7da457d2ba1f3c2f01c7c9bf208e52be36/x/authz/spec/03_messages.md#MsgRevoke) for revoking an existing grant
 - [MsgExec](https://github.com/cosmos/cosmos-sdk/blob/75f5dd7da457d2ba1f3c2f01c7c9bf208e52be36/x/authz/spec/03_messages.md#MsgExec) for executing a message if there exists a grant allowing the sender of the message to do so.
 
 Each grant can be equipped with the expiration time (when it expires, the grant is no longer valid).

## TLA+ specification description
The state of our `authz` model is defined by two variables (in the file `authz.tla`):
- `active_grants`, which is a function (a mapping) from grants (granter, grantee, sdk_message_type) to grant payloads
- `expired_grants`, which is a set of grants that are expired

There are additionaly variables used to describe different testing scenarios (e.g., why the execution failed, or how many grants have been given altogether).

The transitions of the model are defined by four predicates:
- `GiveGrant(grant, grant_payload)`: Creating a new grant `grant` with the payload `grant_payload`. This is modelled by updating the function `active_grants`. Giving a grant fails (and the predicate returns `False`) if granter and grantee have the same address 

- `RevokeGrant(grant)`: Revoking an existing grant `grant`. The predicate returns `False` if there is no grant `grant` or if granter's and grantee's address is the same.

- `ExpireGrant(grant)`: Reaching the expiration time of the grant is abstracted away by the predicate `ExpireGrant`, which can ocurr independently for any of the active grants. What is lost then is the relation between different grants and their expirations (in real life, there could be a dependency: if A expires, then definitely B and C have to expire). Furthermore, we lose the ability to model rejecting a grant whose expiration time is behind the current time. This action is a pure environment action and thus it does not remove the grant from a set of active grants (the system will do so once it notices that the grant expired, upon running the Execute function).

- `Execute(sender, signer, msg)`: Executing the message `msg` by `sender` on behalf of `signer`.
The predicate checks whether there is an appropriate grant for the execution message, but does not model the actual execution and whether or not the execution itself fails. The main logic of checking whether there is an appropriate grant is implemented in the predicate `Accept(grant, grant_payload, msg)`. This predicate returns a record of type `EXEC_OUTCOME`, which closely follows the [`AcceptResponse`](https://github.com/cosmos/cosmos-sdk/blob/8cfa2c26738276d895caf9eb98b3f70616218e17/x/authz/authorizations.go#L29) struct from `authz` (this is new, after implementing a `go` driver, when it proved useful to stay closer to the code).

## Files

Model:
- [`authz.tla`](authz.tla): the model, with 4 transition predicates described above.
- [`helper_predicates.tla`](helper_predicates.tla): contains helper predicates. (They would fit into `authz.tla` too, but would hinder readability). 
- [`generic_authz.tla`](generic_authz.tla), [`stake_authz.tla`](stake_authz.tla), [`send_authz.tla`](send_authz.tla): predicates specific to different authorization types.
- [`authz_typedefs.tla`](authz_typedefs.tla): Apache types, as well as sets of all messages, payloads etc., and also predicates restricting certain supertypes for specific use-cases (e.g., defining that a `send` authorization should have `allow_list` set to its default value, as it is a field necessary only for `stake` authorizations)
-[`statusCodes.tla`](statusCodes.tla): different reserved values (used to avoid writing them as constants throughout other files)

Tests:
- [`authz_test.tla`](authz_test.tla): initial conditions and test cases (negated invariants) for the `authz` module (permiting all authorization types)
- [`authz_test.cfg`](authz_test.cfg): configuration file, defining constants and invariants to check
- [`stake_authz_test.tla`](stake_authz_test.tla): some tests that are specific to staking authorizations
- [`stake_authz_test.cfg`](stake_authz_test.cfg), [`generic_authz_test.cfg`](generic_authz_test.cfg), [`stake_authz_test.cfg`](stake_authz_test.cfg): configuration files testing individual authorization types


## Open questions
1. Perhaps it would be meaningful to explicitly model all failures by a special failure state (instead of forbidding those situation in the model, as it is done now). That way, more interesting behavior could be tested
2. Some test cases also have its _complexity component_. For instance, a test may require that a grant is successfully given, but also that there are at least 3 active grants already existing. These complexity components are orthogonal to the actual tested predicate and they try to introduce a bit more interaction (which could in a real system reveal more bugs). It would be nice to figure out how to factor out those components and include them separately (currently, they are baked into predicates).
