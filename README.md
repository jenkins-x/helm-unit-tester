# helm-unit-tester

a small golang library for unit testing helm charts from go code

## Example

To see an example of how you can create some unit tests on a helm chart with different input values see [this test case](pkg/asserts_test.go#L11-L14)

Then to create your test cases creaate a [tests directory](pkg/test_data/tests). 

For each test directory you can give a list of values YAML files to be passed into `helm template` to override the default values in the chart like [this example](pkg/test_data/tests/knative/values).

Then you can specify the folder tree of the expected resources created your chart such as [this example](pkg/test_data/tests/default/expected/apps/v1/Deployment).

You can also run your test and then just copy the generated yaml files into your `expected` folder ;) 