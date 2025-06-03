package main

import (
	"fmt"
	"os"
	"strings"

	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/spf13/cobra"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"

	// If using ginkgo, import your tests here
	_ "github.com/openshift/operator-framework-operator-controller/test/origin-extension/test/extended/olmv1dev"
	_ "github.com/openshift/operator-framework-operator-controller/test/origin-extension/test/extended/olmv1qe"
)

func main() {
	// Extension registry
	registry := e.NewRegistry()

	// You can declare multiple extensions, but most people will probably only need to create one.
	ext := e.NewExtension("openshift", "payload", "olmv1")                     // the specs of this ext will be ran by openshift-tests
	extNonDefault := e.NewExtension("openshift", "payload", "olmv1NonDefault") // the specs of this ext will be not ran by openshift-tests

	ext.AddSuite(
		e.Suite{
			Name:    "olmv1/parallel",
			Parents: []string{"openshift/conformance/parallel"},
			Qualifiers: []string{
				`!(name.contains("[Serial]") || name.contains("[Disruptive]")  || name.contains("[Slow]"))`,
			},
		})

	ext.AddSuite(
		e.Suite{
			Name:    "olmv1/allparallel",
			Qualifiers: []string{
				`!(name.contains("[Serial]") || name.contains("[Disruptive]"))`,
			},
		})

	ext.AddSuite(e.Suite{
		Name: "olmv1/serial",
		Qualifiers: []string{
			`labels.exists(l, l=="SERIAL")`,
		},
	})

	ext.AddSuite(e.Suite{
		Name: "olmv1/slow",
		Qualifiers: []string{
			`labels.exists(l, l=="SLOW")`,
		},
	})

	// If using Ginkgo, build test specs automatically
	specs, err := g.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		panic(fmt.Sprintf("couldn't build extension test specs from ginkgo: %+v", err.Error()))
	}
	specs.Select(et.NameContains("[Slow]")).AddLabel("SLOW")
	specs.SelectAny([]et.SelectFunction{et.NameContains("[Serial]"), et.NameContains("[Disruptive]")}).AddLabel("SERIAL")

	defaultSpecs := specs.Select(NameDontContains("[NONDEFAULT]"))
	nonDefaultSpecs := specs.Select(et.NameContains("[NONDEFAULT]"))
	// You can add hooks to run before/after tests. There are BeforeEach, BeforeAll, AfterEach,
	// and AfterAll. "Each" functions must be thread safe.
	//
	// specs.AddBeforeAll(func() {
	// 	initializeTestFramework()
	// })
	//
	// specs.AddBeforeEach(func(spec ExtensionTestSpec) {
	//	if spec.Name == "my test" {
	//		// do stuff
	//	}
	// })
	//
	// specs.AddAfterEach(func(res *ExtensionTestResult) {
	// 	if res.Result == ResultFailed && apiTimeoutRegexp.Matches(res.Output) {
	// 		res.AddDetails("api-timeout", collectDiagnosticInfo())
	// 	}
	// })

	// You can also manually build a test specs list from other testing tooling
	// TODO: example

	// Modify specs, such as adding a label to all specs
	// 	specs = specs.AddLabel("SLOW")

	// Specs can be globally filtered...
	// specs = specs.MustFilter([]string{`name.contains("filter")`})

	// Or walked...
	// specs = specs.Walk(func(spec *extensiontests.ExtensionTestSpec) {
	//	if strings.Contains(e.Name, "scale up") {
	//		e.Labels.Insert("SLOW")
	//	}
	//
	// Specs can also be selected...
	// specs = specs.Select(et.NameContains("slow test")).AddLabel("SLOW")
	//
	// Or with "any" (or) matching selections
	// specs = specs.SelectAny(et.NameContains("slow test"), et.HasLabel("SLOW"))
	//
	// Or with "all" (and) matching selections
	// specs = specs.SelectAll(et.NameContains("slow test"), et.HasTagWithValue("speed", "slow"))
	//
	// There are also Must* functions for any of the above flavors of selection
	// which will return an error if nothing is found
	// specs, err = specs.MustSelect(et.NameContains("slow test")).AddLabel("SLOW")
	// if err != nil {
	//    logrus.Warn("no specs found: %w", err)
	// }
	// Test renames
	//	if spec.Name == "[sig-testing] openshift-tests-extension has a test with a typo" {
	//		spec.OriginalName = `[sig-testing] openshift-tests-extension has a test with a tpyo`
	//	}
	//
	// Filter by environment flags
	// if spec.Name == "[sig-testing] openshift-tests-extension should support defining the platform for tests" {
	//		spec.Include(et.PlatformEquals("aws"))
	//		spec.Exclude(et.And(et.NetworkEquals("ovn"), et.TopologyEquals("ha")))
	//	}
	// })

	ext.AddSpecs(defaultSpecs)
	extNonDefault.AddSpecs(nonDefaultSpecs)
	registry.Register(ext)
	registry.Register(extNonDefault)

	root := &cobra.Command{
		Long: "OpenShift Tests Extension Example",
	}

	root.AddCommand(cmd.DefaultExtensionCommands(registry)...)

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		os.Exit(1)
	}
}

func NameDontContains(name string) et.SelectFunction {
	return func(spec *et.ExtensionTestSpec) bool {
		return !strings.Contains(spec.Name, name)
	}
}
