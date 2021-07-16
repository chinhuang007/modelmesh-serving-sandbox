---
title: Testing
---

This repository contains two types of tests:

- Unit Tests which leverage the golang test facilities to perform unit evaluation of functions
- Integration tests which evaluate reconciler execution using a Kubernetes API server instance

## Test Frameworks

Golang includes test capabilities with the core language SDK. You can read about this [here](https://gobyexample.com/testing). Within this repository, this is the preferred framework for _unit tests_.

There are many other test frameworks for golang. One of the more popular frameworks is Ginkgo which is a BDD test framework. Ginkgo adds functions common in test libraries found in other languages such as matchers, pre-post hooks for setup and teardown and other functions. However, it also adds some complexity and overhead, so we do not use this in all of our tests.

The kubebuilder toolset generates Ginkgo test scaffolds for reconcilers as well as a short lived Kubernetes API Server instance which comes from the controller-runtime library. This is known as `envtest`. Ginkgo is only used within the controller project where the reconcilers are tested using the api server.

## Testing Controllers/Reconcilers

The reconcilers found in the 'controllers' package have a few characteristcs of note:

1. They are authored with the Ginkgo BDD style using 'Describe', 'Context', 'It' verbs
2. There is a BeforeSuite hook which is found in the [suite_test.go](controllers/suite_test.go) file which starts a Kubernetes API Server instance using envtest prior to the test execution
3. The tests in this package may use snapshots to evaluate produced objects against previously generated results

## Snapshots

Snapshots allow a test to capture an object state and save it in source control for future comparison. On the first execution, the snapshot file will be saved into the 'testdata' directory as a subdirectory of the currently executing test. When the test is later executed, the current object state will be compared against the snapshot to assure that the object structure as not inadvertently changed. This is effectively a substitute for a series of test matching checks which compares expected output with current output.

### Handling Snapshot Failures

If a functional change causes a different structure to be produced from that of the original snapshot, you will see a test failure which describes the difference in output. There are a few common reasons that this will occur:

1. The structure was inadvertently changed
2. The prior structure found in the snapshot was incorrect and has been fixed

In the first case, you should work through your modifications and determine the reason that the snapshot was altered. Although deleting the snapshot will cause it to be regenerated and allow the test to pass, care must be taken doing so as this will effectively disable the test.

In the latter case, you should either edit the snapshot or remove it so that it can be regenerated. When you check in the functional change, this snapshot update should also be included in your PR.

### Reviewer Responbilities

If you are reviewing a change which includes a snapshot update, you should inspect the snapshot to assure that the new structural modification is desired.
