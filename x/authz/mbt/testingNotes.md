- It is not clear to me why `generic_authorization.go` returns `nil` for the updated authorization.
 The behaviour I expected was that it should return the same authorization (the update does not happen)
Note: I accounted for it in the driver's code, but perhaps that should be an error?

- It is not clear to me why `staking/types/authz.go` never puts `False` under reponse.Accept, but 
rather throws an error. I did not account for it in the driver's code and am treating it like
an error (perhaps it is not?)

- Something seems to be off in my understanding of `denyList`. 
    - Claim1: exactly one of `allowList` and `denyList` may be given (this follows from the function `validateAndBech32fy`)
    - Claim2: delegation will be accepted if the validator is both in the `allowList` and not in the `denyList` (this follows from the function `Accept` of `authz.go`: the boolean value `isValidatorExists` has to be set to `true`, or otherwise an error is raised)
    - Conclusion: if the authorization is to ever be successfuly used, `denyList` must always be empty

- staking implementation of Accept uses function `Sub` which panics if a negative amount is a result. This is then not handled properly in `staking/types/authz.go`. For the reference, the same logic in `x/bank` is implemented using `SafeSub`

- in the function `ValidateBasic` of `x/bank/types/send_authorization.go` the error message is "spend limit cannot be negative", when in fact it also rejects 0. A better error message would be "... must be positive". Additionally, this is missing from the documentation

- it does not seem that the `SaveGrant` of `authz/keeper` checks that `granter` does not equal `grantee` (as is specified in the English spec). Is it checked at some other level, or is this a bug? (The same question for `DeleteGrant`.)

