# Contributing to Terraform - Active Directory Provider

**First:** if you're unsure or afraid of _anything_, ask for help! You can
submit a work in progress (WIP) pull request, or file an issue with the parts
you know. We'll do our best to guide you in the right direction, and let you
know if there are guidelines we will need to follow. We want people to be able
to participate without fear of doing the wrong thing.

Below are our expectations for contributors. Following these guidelines gives us
the best opportunity to work with you, by making sure we have the things we need
in order to make it happen. Doing your best to follow it will speed up our
ability to merge PRs and respond to issues.

<!-- TOC depthFrom:2 -->

- [Issues](#issues)
    - [Issue Reporting Checklists](#issue-reporting-checklists)
        - [Bug Reports](#bug-reports)
        - [Feature Requests](#feature-requests)
        - [Questions](#questions)
    - [Issue Lifecycle](#issue-lifecycle)
- [Pull Requests](#pull-requests)
    - [Pull Request Lifecycle](#pull-request-lifecycle)
    - [Checklists for Contribution](#checklists-for-contribution)
        - [Documentation Update](#documentation-update)
        - [Enhancement/Bugfix to a Resource](#enhancementbugfix-to-a-resource)
        - [New Resource](#new-resource)
    - [Common Review Items](#common-review-items)
        - [Go Coding Style](#go-coding-style)
        - [Resource Contribution Guidelines](#resource-contribution-guidelines)
        - [Acceptance Testing Guidelines](#acceptance-testing-guidelines)
    - [Writing Acceptance Tests](#writing-acceptance-tests)
        - [Acceptance Tests Often Cost Money to Run](#acceptance-tests-often-cost-money-to-run)
        - [Running an Acceptance Test](#running-an-acceptance-test)
        - [Writing an Acceptance Test](#writing-an-acceptance-test)

<!-- /TOC -->

## Issues

### Issue Reporting Checklists

We welcome issues of all kinds including feature requests, bug reports, and
general questions. Below you'll find checklists with guidelines for well-formed
issues of each type.

#### [Bug Reports](https://github.com/adlerrobert/terraform-provider-activedirectory/issues/new?template=---bug-report.md)

 - [ ] __Test against latest release__: Make sure you test against the latest
   released version. It is possible we already fixed the bug you're experiencing.

 - [ ] __Search for possible duplicate reports__: It's helpful to keep bug
   reports consolidated to one thread, so do a quick search on existing bug
   reports to check if anybody else has reported the same thing. You can [scope
      searches by the label "bug"](https://github.com/adlerrobert/terraform-provider-activedirectory/issues?q=is%3Aopen+is%3Aissue+label%3Abug) to help narrow things down.

 - [ ] __Include steps to reproduce__: Provide steps to reproduce the issue,
   along with your `.tf` files, with secrets removed, so we can try to
   reproduce it. Without this, it makes it much harder to fix the issue.

 - [ ] __For panics, include `crash.log`__: If you experienced a panic, please
   create a [gist](https://gist.github.com) of the *entire* generated crash log
   for us to look at. Double check no sensitive items were in the log.

#### [Feature Requests](https://github.com/adlerrobert/terraform-provider-activedirectory/issues/new?labels=enhancement&template=---feature-request.md)

 - [ ] __Search for possible duplicate requests__: It's helpful to keep requests
   consolidated to one thread, so do a quick search on existing requests to
   check if anybody else has reported the same thing. You can [scope searches by
      the label "enhancement"](https://github.com/adlerrobert/terraform-provider-activedirectory/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement) to help narrow things down.

 - [ ] __Include a use case description__: In addition to describing the
   behavior of the feature you'd like to see added, it's helpful to also lay
   out the reason why the feature would be important and how it would benefit
   Terraform users.

#### [Questions](https://github.com/adlerrobert/terraform-provider-activedirectory/issues/new?labels=question&template=---question.md)

 - [ ] __Search for answers in Terraform documentation__: We're happy to answer
   questions in GitHub Issues. Oftentimes Question issues result in documentation updates
   to help future users, so if you don't find an answer, you can give us
   pointers for where you'd expect to see it in the docs.

### Issue Lifecycle

1. The issue is reported.

2. The issue is verified and categorized.
   Categorization is done via GitHub labels.

3. An initial process determines whether the issue is critical and must
    be addressed immediately, or can be left open for community discussion.

4. The issue is addressed in a pull request or commit. The issue number will be
   referenced in the commit message so that the code that fixes it is clearly
   linked.

5. The issue is closed. Sometimes, valid issues will be closed because they are
   tracked elsewhere or non-actionable. The issue is still indexed and
   available for future viewers, or can be re-opened if necessary.

## Pull Requests

We appreciate direct contributions to the provider codebase. Here's what to
expect:

 * For pull requests that follow the guidelines, we will proceed to reviewing
  and merging, following the provider team's review schedule. There may be some
  internal or community discussion needed before we can complete this.
 * Pull requests that don't follow the guidelines will be commented with what
  they're missing. The person who submits the pull request or another community
  member will need to address those requests before they move forward.

### Pull Request Lifecycle

1. [Fork the GitHub repository](https://help.github.com/en/articles/fork-a-repo),
   modify the code, and [create a pull request](https://help.github.com/en/articles/creating-a-pull-request-from-a-fork).
   You are welcome to submit your pull request for commentary or review before
   it is fully completed by creating a [draft pull request](https://help.github.com/en/articles/about-pull-requests#draft-pull-requests)
   or adding `[WIP]` to the beginning of the pull request title.
   Please include specific questions or items you'd like feedback on.

1. Once you believe your pull request is ready to be reviewed, ensure the
   pull request is not a draft pull request by [marking it ready for review](https://help.github.com/en/articles/changing-the-stage-of-a-pull-request)
   or removing `[WIP]` from the pull request title if necessary, and a
   maintainer will review it. Follow [the checklists below](#checklists-for-contribution)
   to help ensure that your contribution can be easily reviewed and potentially
   merged.

1. One of team members will look over your contribution and
   either approve it or provide comments letting you know if there is anything
   left to do.

1. Once all outstanding comments and checklist items have been addressed, your
   contribution will be merged! Merged PRs will be included in the next
   release. The provider team takes care of updating the CHANGELOG as
   they merge.

1. In some cases, we might decide that a PR should be closed without merging.
   We'll make sure to provide clear reasoning when this happens.

### Checklists for Contribution

There are several different kinds of contribution, each of which has its own
standards for a speedy review. The following sections describe guidelines for
each type of contribution.

#### Documentation Update (WIP)

#### Enhancement/Bugfix to a Resource

Working on existing resources is a great way to get started as a Terraform
contributor because you can work within existing code and tests to get a feel
for what to do.

In addition to the below checklist, please see the [Common Review
Items](#common-review-items) sections for more specific coding and testing
guidelines.

 - [ ] __Acceptance test coverage of new behavior__: Existing resources each
   have a set of [acceptance tests][acctests] covering their functionality.
   These tests should exercise all the behavior of the resource. Whether you are
   adding something or fixing a bug, the idea is to have an acceptance test that
   fails if your code were to be removed. Sometimes it is sufficient to
   "enhance" an existing test by adding an assertion or tweaking the config
   that is used, but it's often better to add a new test. You can copy/paste an
   existing test and follow the conventions you see there, modifying the test
   to exercise the behavior of your code.
 - [ ] __Documentation updates (WIP)__: If your code makes any changes that need to
   be documented, you should include those doc updates in the same PR. This
   includes things like new resource attributes or changes in default values.
   The [Terraform website][website] source is in this repo and includes
   instructions for getting a local copy of the site up and running if you'd
   like to preview your changes.
 - [ ] __Well-formed Code__: Do your best to follow existing conventions you
   see in the codebase, and ensure your code is formatted with `go fmt`. (The
   Circle CI build will fail if `go fmt` has not been run on incoming code.)
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
 - [ ] __Vendor additions__: Create a separate PR if you are updating the vendor
   folder. This is to avoid conflicts as the vendor versions tend to be fast-
   moving targets. We will plan to merge the PR with this change first.

#### New Resource

Implementing a new resource is a good way to learn more about how Terraform
interacts with upstream APIs. There are plenty of examples to draw from in the
existing resources, but you still get to implement something completely new.

In addition to the below checklist, please see the [Common Review
Items](#common-review-items) sections for more specific coding and testing
guidelines.

 - [ ] __Minimal LOC__: It's difficult for both the reviewer and author to go
   through long feedback cycles on a big PR with many resources. We ask you to
   only submit **1 resource at a time**.
 - [ ] __Acceptance tests__: New resources should include acceptance tests
   covering their behavior. See [Writing Acceptance
   Tests](#writing-acceptance-tests) below for a detailed guide on how to
   approach these.
 - [ ] __Resource Naming__: Resources should be named `activedirectory_<name>`,
   using underscores (`_`) as the separator.
 - [ ] __Arguments_and_Attributes__: The HCL for arguments and attributes should
   mimic the types and structs presented by the Active Directory API. API arguments should be
   converted from `CamelCase` to `camel_case`.
 - [ ] __Documentation__: - need to be done
 - [ ] __Well-formed Code__: Do your best to follow existing conventions you
   see in the codebase, and ensure your code is formatted with `go fmt`. (The
   Travis CI build will fail if `go fmt` has not been run on incoming code.)
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
 - [ ] __Vendor updates__: Create a separate PR if you are adding to the vendor
   folder. This is to avoid conflicts as the vendor versions tend to be fast-
   moving targets. We will plan to merge the PR with this change first.

### Common Review Items

The Terraform Active Directory Provider follows common practices to ensure consistent and
reliable implementations across all resources in the project. While there may be
older resource and testing code that predates these guidelines, new submissions
are generally expected to adhere to these items to maintain Terraform Provider
quality. For any guidelines listed, contributors are encouraged to ask any
questions and community reviewers are encouraged to provide review suggestions
based on these guidelines to speed up the review and merge process.

#### Go Coding Style

The following Go language resources provide common coding preferences that may be referenced during review, if not automatically handled by the project's linting tools.

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

#### Resource Contribution Guidelines

The following resource checks need to be addressed before your contribution can be merged. The exclusion of any applicable check may result in a delayed time to merge.

- [ ] __Passes Testing__: All code and documentation changes must pass unit testing, code linting, and website link testing. Resource code changes must pass all acceptance testing for the resource.
- [ ] __Avoids Optional and Required for Non-Configurable Attributes__: Resource schema definitions for read-only attributes should not include `Optional: true` or `Required: true`.
- [ ] __Avoids resource.Retry() without resource.RetryableError()__: Resource logic should only implement [`resource.Retry()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#Retry) if there is a retryable condition (e.g. `return resource.RetryableError(err)`).
- [ ] __Avoids Resource Read Function in Data Source Read Function__: Data sources should fully implement their own resource `Read` functionality including duplicating `d.Set()` calls.
- [ ] __Avoids Reading Schema Structure in Resource Code__: The resource `Schema` should not be read in resource `Create`/`Read`/`Update`/`Delete` functions to perform looping or otherwise complex attribute logic. Use [`d.Get()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Get) and [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) directly with individual attributes instead.
- [ ] __Avoids ResourceData.GetOkExists()__: Resource logic should avoid using [`ResourceData.GetOkExists()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.GetOkExists) as its expected functionality is not guaranteed in all scenarios.
- [ ] __Implements Read After Create and Update__: Except where API eventual consistency prohibits immediate reading of resources or updated attributes,  resource `Create` and `Update` functions should return the resource `Read` function.
- [ ] __Implements Immediate Resource ID Set During Create__: Immediately after calling the API creation function, the resource ID should be set with [`d.SetId()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.SetId) before other API operations or returning the `Read` function.
- [ ] __Implements Attribute Refreshes During Read__: All attributes available in the API should have [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) called their values in the Terraform state during the `Read` function.
- [ ] __Implements Error Checks with Non-Primative Attribute Refreshes__: When using [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) with non-primative types (`schema.TypeList`, `schema.TypeSet`, or `schema.TypeMap`), perform error checking to [prevent issues where the code is not properly able to refresh the Terraform state](https://www.terraform.io/docs/extend/best-practices/detecting-drift.html#error-checking-aggregate-types).
- [ ] __Implements Import Acceptance Testing and Documentation (WIP)__:
- [ ] __Implements Customizable Timeouts Documentation__: Support for customizable timeouts (`Timeouts` in resource schema) must include `## Timeouts` section in resource documentation.
- [ ] __Implements State Migration When Adding New Virtual Attribute__: For new "virtual" attributes (those only in Terraform and not in the API), the schema should implement [State Migration](https://www.terraform.io/docs/extend/resources.html#state-migrations) to prevent differences for existing configurations that upgrade.
- [ ] __Uses TypeList and MaxItems: 1__: Configuration block attributes (e.g. `Type: schema.TypeList` or `Type: schema.TypeSet` with `Elem: &schema.Resource{...}`) that can only have one block should use `Type: schema.TypeList` and `MaxItems: 1` in the schema definition.
- [ ] __Uses Existing Validation Functions__: Schema definitions including `ValidateFunc` for attribute validation should use available [Terraform `helper/validation` package](https://godoc.org/github.com/hashicorp/terraform/helper/validation) functions. `All()`/`Any()` can be used for combining multiple validation function behaviors.
- [ ] __Uses resource.NotFoundError__: Custom errors for missing resources should use [`resource.NotFoundError`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#NotFoundError).
- [ ] __Skips Exists Function__: Implementing a resource `Exists` function is extraneous as it often duplicates resource `Read` functionality. Ensure `d.SetId("")` is used to appropriately trigger resource recreation in the resource `Read` function.
- [ ] __Skips id Attribute__: The `id` attribute is implicit for all Terraform resources and does not need to be defined in the schema.

The below are style-based items that _may_ be noted during review and are recommended for simplicity, consistency, and quality assurance:

- [ ] __Avoids CustomizeDiff__: Usage of `CustomizeDiff` is generally discouraged.
- [ ] __Implements Error Message Context (WIP)__: Returning errors from resource `Create`, `Read`, `Update`, and `Delete` functions should include additional messaging about the location or cause of the error for operators and code maintainers by wrapping with [`fmt.Errorf()`](https://godoc.org/golang.org/x/exp/errors/fmt#Errorf).
  - An example `Delete` API error: `return fmt.Errorf("error deleting {THING} (%s): %s", d.Id(), err)`
  - An example `d.Set()` error: `return fmt.Errorf("error setting {ATTRIBUTE}: %s", err)`
- [ ] __Implements Warning Logging With Resource State Removal (WIP)__: If a resource is removed outside of Terraform (e.g. via different tool, API, or web UI), `d.SetId("")` and `return nil` can be used in the resource `Read` function to trigger resource recreation. When this occurs, a warning log message should be printed beforehand: `log.Printf("[WARN] {THING} (%s) not found, removing from state", d.Id())`
- [ ] __Uses Elem with TypeMap__: While provider schema validation does not error when the `Elem` configuration is not present with `Type: schema.TypeMap` attributes, including the explicit `Elem: &schema.Schema{Type: schema.TypeString}` is recommended.
- [ ] __Uses American English for Attribute Naming__: For any ambiguity with attribute naming, prefer American English over British English. e.g. `color` instead of `colour`.
- [ ] __Skips Timestamp Attributes__: Generally, creation and modification dates from the API should be omitted from the schema.
- [ ] __Skips Error() Call__: Error objects do not need to have `Error()` called.

#### Acceptance Testing Guidelines

The below are required items that will be noted during submission review and prevent immediate merging:

- [ ] __Implements CheckDestroy__: Resource testing should include a `CheckDestroy` function (typically named `testAccCheckAD{RESOURCE}Destroy`) that calls the API to verify that the Terraform resource has been deleted or disassociated as appropriate. More information about `CheckDestroy` functions can be found in the [Extending Terraform TestCase documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy).
- [ ] __Implements Exists Check Function__: Resource testing should include a `TestCheckFunc` function (typically named `testAccCheckAD{RESOURCE}Exists`) that calls the API to verify that the Terraform resource has been created or associated as appropriate. Preferably, this function will also accept a pointer to an API object representing the Terraform resource from the API response that can be set for potential usage in later `TestCheckFunc`. More information about these functions can be found in the [Extending Terraform Custom Check Functions documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy).
- [ ] __Excludes Provider Declarations__: Test configurations should not include `provider "activedirectory" {...}` declarations. If necessary, only the provider declarations in `provider_test.go` should be used for multiple account/region or otherwise specialized testing.
- [ ] __Uses resource.ParallelTest (WIP)__: Tests should utilize [`resource.ParallelTest()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#ParallelTest) instead of [`resource.Test()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#Test) except where serialized testing is absolutely required.
- [ ] __Uses fmt.Sprintf()__: Test configurations preferably should to be separated into their own functions (typically named `testAccADRESOURCE}Config{PURPOSE}`) that call [`fmt.Sprintf()`](https://golang.org/pkg/fmt/#Sprintf) for variable injection or a string `const` for completely static configurations. Test configurations should avoid `var` or other variable injection functionality such as [`text/template`](https://golang.org/pkg/text/template/).
- [ ] __Uses Randomized Infrastructure Naming (WIP)__: Test configurations that utilize resources where a unique name is required should generate a random name. Typically this is created via `rName := acctest.RandomWithPrefix("tf-acc-test")` in the acceptance test function before generating the configuration.

### Writing Acceptance Tests

Terraform includes an acceptance test harness that does most of the repetitive
work involved in testing a resource. For additional information about testing
Terraform Providers, see the [Extending Terraform documentation](https://www.terraform.io/docs/extend/testing/index.html).

#### Acceptance Tests Often Cost Money to Run

Because acceptance tests create real resources, they often cost money to run.
Because the resources only exist for a short period of time, the total amount
of money required is usually a relatively small.

#### Running an Acceptance Test

Acceptance tests can be run using the `testacc` target in the Terraform
`Makefile`. The individual tests to run can be controlled using a regular
expression. Prior to running the tests provider configuration details such as
access keys must be made available as environment variables.

For example, to run an acceptance test against the Active Directory
provider, the following environment variables must be set:

```sh
export AD_HOST=...
export AD_BIND_USER=...
export AD_BIND_PASSWORD=...
export AD_COMPUTER_TEST_BASE_OU=...
```

#### Writing an Acceptance Test

Terraform has a framework for writing acceptance tests which minimises the
amount of boilerplate code necessary to use common testing patterns. The entry
point to the framework is the `resource.ParallelTest()` function.

Tests are divided into `TestStep`s. Each `TestStep` proceeds by applying some
Terraform configuration using the provider under test, and then verifying that
results are as expected by making assertions using the provider API. It is
common for a single test function to exercise both the creation of and updates
to a single resource. Most tests follow a similar structure.

1. Pre-flight checks are made to ensure that sufficient provider configuration
   is available to be able to proceed - for example in an acceptance test
   targeting AD, `AD_HOST`, `AD_BIND_USER`, `AD_BIND_PASSWORD` and `AD_COMPUTER_TEST_BASE_OU` must be set prior
   to running acceptance tests. This is common to all tests exercising a single
   provider.

Each `TestStep` is defined in the call to `resource.ParallelTest()`. Most assertion
functions are defined out of band with the tests. This keeps the tests
readable, and allows reuse of assertion functions across different tests of the
same type of resource. The definition of a complete test looks like this:

```go
func TestAccADComputer_basic(t *testing.T) {
	ou := os.Getenv("AD_COMPUTER_TEST_BASE_OU")
	name := "test-acc-computer"
	description := "terraform"

	var computer Computer

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckADComputerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceADComputerTestData(ou, name, description),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckADComputerExists("activedirectory_computer.test", &computer),
					testAccCheckADComputerAttributes(&computer, ou, name, description),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "ou", ou),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "name", name),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "description", description),
					resource.TestCheckResourceAttr("activedirectory_computer.test", "id", fmt.Sprintf("cn=%s,%s", name, ou)),
				),
			},
		},
	})
}
```

[website]: https://github.com/adlerrobert/terraform-provider-activedirectory/tree/master/website
[acctests]: https://github.com/hashicorp/terraform#acceptance-tests
