# helm-unit-tester

a small golang library for unit testing helm charts from go code

## Example

To see an example of how you can create some unit tests on a helm chart with different input values see [this test case](pkg/asserts_test.go#L11-L14)

Then to create your test cases creaate a [tests directory](pkg/test_data/tests). 

For each test directory you can give a list of values YAML files to be passed into `helm template` to override the default values in the chart like [this example](pkg/test_data/tests/knative/values).

Then you can specify the folder tree of the expected resources created your chart such as [this example](pkg/test_data/tests/default/expected/apps/v1/Deployment).

You can also run your test and then just copy the generated yaml files into your `expected` folder ;)

### testcase.yml

You can also override the default properties of each `TestCase` via a little [testcase.yml](pkg/test_data/tests/missing-resource/testcase.yml) file in your test case if you want to. e.g. to enable diferent flags.

## Helm 2 v 3

The library handles the CLI differences between 2 and 3; though note ther are subtle differences in the output of 2 and 3

## Regenerating the expected YAML files

sometimes you'll change helm versions or change the underlying charts and the YAML that is generated will change.

To auto-regenerate the expected YAML you can use a feature flag

``` 
export HELM_UNIT_REGENERATE_EXPECTED=true
```

then run your tests and the expected yamls will be regenerated for you. You can do a git diff to view the actual changes.

## Adding custom validation

Once the unit tests have completed you may want to add your own validation to verify the generated YAML looks how you imagine it should.

To do that just iterate over the `TestCase` results and parse the YAML and verify however you like in go code.

Here is an [example test case doing that](https://github.com/jenkins-x-charts/jxboot-helmfile-resources/blob/master/tests/chart_test.go#L13) which verifies all kinds of different input configurations (e.g. for custom environments, no environments, remote environments and whatnot)
