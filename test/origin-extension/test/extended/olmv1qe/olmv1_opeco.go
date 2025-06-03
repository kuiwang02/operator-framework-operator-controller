package olmv1qe

import (
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	olmv1util "github.com/openshift/operator-framework-operator-controller/test/origin-extension/test/extended/olmv1qe/olmv1util"
	exutil "github.com/openshift/operator-framework-operator-controller/test/origin-extension/test/extended/util"
)

var _ = g.Describe("[sig-operators] OLM v1 opeco should", func() {
	// Hypershift will be supported from 4.19, so add NonHyperShiftHOST Per cases now.
	defer g.GinkgoRecover()
	var (
		oc = exutil.NewCLI("olmv1-opeco"+getRandomString(), exutil.KubeConfigPath())
	)

	g.BeforeEach(func() {
		exutil.SkipNoOLMv1Core(oc)
	})

	// author: xzha@redhat.com
	g.It("Author:xzha-LEVEL0-ROSA-OSD_CCS-ARO-ConnectedOnly-NonHyperShiftHOST-High-69123-Catalogd clustercatalog offer the operator content through http server", func() {
		var (
			baseDir                = exutil.FixturePath("testdata", "olm", "v1")
			clustercatalogTemplate = filepath.Join(baseDir, "clustercatalog.yaml")
			clustercatalog         = olmv1util.ClusterCatalogDescription{
				Name:     "clustercatalog-69123",
				Imageref: "quay.io/olmqe/olmtest-operator-index:nginxolm69123",
				Template: clustercatalogTemplate,
			}
		)
		exutil.By("Create clustercatalog")
		defer clustercatalog.Delete(oc)
		clustercatalog.Create(oc)

		exutil.By("get the index content through http service on cluster")
		unmarshalContent, err := clustercatalog.UnmarshalContent(oc, "all")
		o.Expect(err).NotTo(o.HaveOccurred())

		allPackageName := olmv1util.ListPackagesName(unmarshalContent.Packages)
		o.Expect(allPackageName[0]).To(o.ContainSubstring("nginx69123"))

		channelData := olmv1util.GetChannelByPakcage(unmarshalContent.Channels, "nginx69123")
		o.Expect(channelData[0].Name).To(o.ContainSubstring("candidate-v0.0"))

		bundlesName := olmv1util.GetBundlesNameByPakcage(unmarshalContent.Bundles, "nginx69123")
		o.Expect(bundlesName[0]).To(o.ContainSubstring("nginx69123.v0.0.1"))

	})

	// author: xzha@redhat.com
	g.It("Author:xzha-ROSA-OSD_CCS-ARO-ConnectedOnly-NonHyperShiftHOST-High-73219-Fetch deprecation data from the catalogd http server [NONDEFAULT]", func() {
		var (
			baseDir                = exutil.FixturePath("testdata", "olm", "v1")
			clustercatalogTemplate = filepath.Join(baseDir, "clustercatalog.yaml")
			clustercatalog         = olmv1util.ClusterCatalogDescription{
				Name:     "clustercatalog-73219",
				Imageref: "quay.io/olmqe/olmtest-operator-index:nginxolm73219",
				Template: clustercatalogTemplate,
			}
		)
		exutil.By("Create clustercatalog")
		defer clustercatalog.Delete(oc)
		clustercatalog.Create(oc)

		exutil.By("get the deprecation content through http service on cluster")
		unmarshalContent, err := clustercatalog.UnmarshalContent(oc, "deprecations")
		o.Expect(err).NotTo(o.HaveOccurred())

		deprecatedChannel := olmv1util.GetDeprecatedChannelNameByPakcage(unmarshalContent.Deprecations, "nginx73219")
		o.Expect(deprecatedChannel[0]).To(o.ContainSubstring("candidate-v0.0"))

		deprecatedBundle := olmv1util.GetDeprecatedBundlesNameByPakcage(unmarshalContent.Deprecations, "nginx73219")
		o.Expect(deprecatedBundle[0]).To(o.ContainSubstring("nginx73219.v0.0.1"))

	})

})
