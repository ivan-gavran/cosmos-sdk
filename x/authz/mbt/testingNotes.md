- It is not clear to me why `generic_authorization.go` returns `nil` for the updated authorization.
I accounted for it in the driver's code, but perhaps that should be an error? The behaviour I expected was that it should return the same authorization (the update does not happen)

- It is not clear to me why `staking/types/authz.go` never puts `False` under reponse.Accept, but 
rather throws an error. I did not account for it in the driver's code and am treating it like
an error (perhaps it is not?)

- Something seems to be off in my understanding of `denyList`. 
    - Claim1: exactly one of `allowList` and `denyList` may be given (this follows from the function `validateAndBech32fy`)
    - Claim2: delegation will be accepted if the validator is both in the `allowList` and not in the `denyList` (this follows from the function `Accept` of `authz.go`: the boolean value `isValidatorExists` has to be set to `true`, or otherwise an error is raised)
    - Conclusion: if the authorization is to ever be successfuly used, `denyList` must always be empty