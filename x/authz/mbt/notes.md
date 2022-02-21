- It is not clear to me why `generic_authorizaion.go` returns `nil` for the updated authorization.
I accounted for it in the driver's code, but perhaps that should be an error?

- It is not clear to me why `staking/types/authz.go` never puts `False` under reponse.Accept, but 
rather throws an error. I did not account for it in the driver's code and am treating it like
an error (perhaps it is not?)