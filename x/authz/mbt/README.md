This folder implements MBT testing for authz.
The subdirectory `model` contains all the information about the model of the system and how to generate traces from it.

Once traces are generated, the relevant formats can be copied to this folder (under `generatedTraces`) using the script from the `utils` folder. (Move to the `utils` folder and run there `python3 moveGeneratedTraces.py`.)

Finally, to run all the tests, use the `authz/mbt_tests.go` (e.g., by running `go test mbt_test.go -v` from the `authz` folder).