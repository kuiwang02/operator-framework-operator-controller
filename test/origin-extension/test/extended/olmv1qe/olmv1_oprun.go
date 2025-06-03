package olmv1qe

import (
	"fmt"
	"path/filepath"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	olmv1util "github.com/openshift/operator-framework-operator-controller/test/origin-extension/test/extended/olmv1qe/olmv1util"
	exutil "github.com/openshift/operator-framework-operator-controller/test/origin-extension/test/extended/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

var _ = g.Describe("[sig-operators] OLM v1 oprun should", func() {
	// Hypershift will be supported from 4.19, so add NonHyperShiftHOST Per cases now.
	defer g.GinkgoRecover()
	var (
		oc = exutil.NewCLIWithoutNamespace("default")
	)

	g.BeforeEach(func() {
		exutil.SkipNoOLMv1Core(oc)
	})

	// author: kuiwang@redhat.com
	g.It("Author:kuiwang-LEVEL0-ROSA-OSD_CCS-ARO-ConnectedOnly-NonHyperShiftHOST-OSD_CCS-Medium-75492-cluster extension can not be installed with wrong sa or insufficient permission sa [NONDEFAULT]", func() {
		exutil.SkipForSNOCluster(oc)
		var (
			caseID                       = "75492"
			ns                           = "ns-" + caseID
			sa                           = "sa" + caseID
			labelValue                   = caseID
			catalogName                  = "clustercatalog-" + caseID
			ceInsufficientName           = "ce-insufficient-" + caseID
			ceWrongSaName                = "ce-wrongsa-" + caseID
			baseDir                      = exutil.FixturePath("testdata", "olm", "v1")
			clustercatalogTemplate       = filepath.Join(baseDir, "clustercatalog-withlabel.yaml")
			clusterextensionTemplate     = filepath.Join(baseDir, "clusterextension-withselectorlabel.yaml")
			saClusterRoleBindingTemplate = filepath.Join(baseDir, "sa-nginx-insufficient-bundle.yaml")
			saCrb                        = olmv1util.SaCLusterRolebindingDescription{
				Name:      sa,
				Namespace: ns,
				RBACObjects: []olmv1util.ChildResource{
					{Kind: "RoleBinding", Ns: ns, Names: []string{fmt.Sprintf("%s-installer-role-binding", sa)}},
					{Kind: "Role", Ns: ns, Names: []string{fmt.Sprintf("%s-installer-role", sa)}},
					{Kind: "ClusterRoleBinding", Ns: "", Names: []string{fmt.Sprintf("%s-installer-rbac-clusterrole-binding", sa),
						fmt.Sprintf("%s-installer-clusterrole-binding", sa)}},
					{Kind: "ClusterRole", Ns: "", Names: []string{fmt.Sprintf("%s-installer-rbac-clusterrole", sa),
						fmt.Sprintf("%s-installer-clusterrole", sa)}},
					{Kind: "ServiceAccount", Ns: ns, Names: []string{sa}},
				},
				Kinds:    "okv3277775492s",
				Template: saClusterRoleBindingTemplate,
			}
			clustercatalog = olmv1util.ClusterCatalogDescription{
				Name:       catalogName,
				Imageref:   "quay.io/olmqe/nginx-ok-index:vokv3283",
				LabelValue: labelValue,
				Template:   clustercatalogTemplate,
			}
			ce75492Insufficient = olmv1util.ClusterExtensionDescription{
				Name:             ceInsufficientName,
				PackageName:      "nginx-ok-v3277775492",
				Channel:          "alpha",
				Version:          ">=0.0.1",
				InstallNamespace: ns,
				SaName:           sa,
				LabelValue:       labelValue,
				Template:         clusterextensionTemplate,
			}
			ce75492WrongSa = olmv1util.ClusterExtensionDescription{
				Name:             ceWrongSaName,
				PackageName:      "nginx-ok-v3277775492",
				Channel:          "alpha",
				Version:          ">=0.0.1",
				InstallNamespace: ns,
				SaName:           sa + "1",
				LabelValue:       labelValue,
				Template:         clusterextensionTemplate,
			}
		)

		exutil.By("Create namespace")
		defer oc.WithoutNamespace().AsAdmin().Run("delete").Args("ns", ns, "--ignore-not-found", "--force").Execute()
		err := oc.WithoutNamespace().AsAdmin().Run("create").Args("ns", ns).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(olmv1util.Appearance(oc, exutil.Appear, "ns", ns)).To(o.BeTrue())

		exutil.By("Create SA for clusterextension")
		defer saCrb.Delete(oc)
		saCrb.Create(oc)

		exutil.By("Create clustercatalog")
		defer clustercatalog.Delete(oc)
		clustercatalog.Create(oc)

		exutil.By("check Insufficient sa from bundle")
		defer ce75492Insufficient.Delete(oc)
		ce75492Insufficient.CreateWithoutCheck(oc)
		if olmv1util.IsFeaturegateEnabled(oc, "NewOLMPreflightPermissionChecks") {
			ce75492Insufficient.CheckClusterExtensionCondition(oc, "Progressing", "message", "pre-authorization failed", 10, 60, 0)
		} else {
			ce75492Insufficient.CheckClusterExtensionCondition(oc, "Progressing", "message", "could not get information about the resource CustomResourceDefinition", 10, 60, 0)
		}
		exutil.By("check wrong sa")
		defer ce75492WrongSa.Delete(oc)
		ce75492WrongSa.CreateWithoutCheck(oc)
		ce75492WrongSa.CheckClusterExtensionCondition(oc, "Installed", "message", "not found", 10, 60, 0)

	})

	// author: kuiwang@redhat.com
	g.It("Author:kuiwang-LEVEL0-ROSA-OSD_CCS-ARO-ConnectedOnly-NonHyperShiftHOST-OSD_CCS-Medium-75493-cluster extension can be installed with enough permission sa", func() {
		exutil.SkipForSNOCluster(oc)
		var (
			caseID                       = "75493"
			ns                           = "ns-" + caseID
			sa                           = "sa" + caseID
			labelValue                   = caseID
			catalogName                  = "clustercatalog-" + caseID
			ceSufficientName             = "ce-sufficient" + caseID
			baseDir                      = exutil.FixturePath("testdata", "olm", "v1")
			clustercatalogTemplate       = filepath.Join(baseDir, "clustercatalog-withlabel.yaml")
			clusterextensionTemplate     = filepath.Join(baseDir, "clusterextension-withselectorlabel.yaml")
			saClusterRoleBindingTemplate = filepath.Join(baseDir, "sa-nginx-limited.yaml")
			saCrb                        = olmv1util.SaCLusterRolebindingDescription{
				Name:      sa,
				Namespace: ns,
				RBACObjects: []olmv1util.ChildResource{
					{Kind: "RoleBinding", Ns: ns, Names: []string{fmt.Sprintf("%s-installer-role-binding", sa)}},
					{Kind: "Role", Ns: ns, Names: []string{fmt.Sprintf("%s-installer-role", sa)}},
					{Kind: "ClusterRoleBinding", Ns: "", Names: []string{fmt.Sprintf("%s-installer-rbac-clusterrole-binding", sa),
						fmt.Sprintf("%s-installer-clusterrole-binding", sa)}},
					{Kind: "ClusterRole", Ns: "", Names: []string{fmt.Sprintf("%s-installer-rbac-clusterrole", sa),
						fmt.Sprintf("%s-installer-clusterrole", sa)}},
					{Kind: "ServiceAccount", Ns: ns, Names: []string{sa}},
				},
				Kinds:    "okv3277775493s",
				Template: saClusterRoleBindingTemplate,
			}
			clustercatalog = olmv1util.ClusterCatalogDescription{
				Name:       catalogName,
				Imageref:   "quay.io/olmqe/nginx-ok-index:vokv3283",
				LabelValue: labelValue,
				Template:   clustercatalogTemplate,
			}
			ce75493 = olmv1util.ClusterExtensionDescription{
				Name:             ceSufficientName,
				PackageName:      "nginx-ok-v3277775493",
				Channel:          "alpha",
				Version:          ">=0.0.1",
				InstallNamespace: ns,
				SaName:           sa,
				LabelValue:       labelValue,
				Template:         clusterextensionTemplate,
			}
		)

		exutil.By("Create namespace")
		defer oc.WithoutNamespace().AsAdmin().Run("delete").Args("ns", ns, "--ignore-not-found", "--force").Execute()
		err := oc.WithoutNamespace().AsAdmin().Run("create").Args("ns", ns).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(olmv1util.Appearance(oc, exutil.Appear, "ns", ns)).To(o.BeTrue())

		exutil.By("Create SA for clusterextension")
		defer saCrb.Delete(oc)
		saCrb.Create(oc)

		exutil.By("Create clustercatalog")
		defer clustercatalog.Delete(oc)
		clustercatalog.Create(oc)

		exutil.By("check if ce is installed with limited permission")
		defer ce75493.Delete(oc)
		ce75493.Create(oc)
		o.Expect(olmv1util.Appearance(oc, exutil.Appear, "customresourcedefinitions.apiextensions.k8s.io", "okv3277775493s.cache.example.com")).To(o.BeTrue())
		o.Expect(olmv1util.Appearance(oc, exutil.Appear, "services", "nginx-ok-v3283-75493-controller-manager-metrics-service", "-n", ns)).To(o.BeTrue())
		ce75493.Delete(oc)
		o.Expect(olmv1util.Appearance(oc, exutil.Disappear, "customresourcedefinitions.apiextensions.k8s.io", "okv3277775493s.cache.example.com")).To(o.BeTrue())
		o.Expect(olmv1util.Appearance(oc, exutil.Disappear, "services", "nginx-ok-v3283-75493-controller-manager-metrics-service", "-n", ns)).To(o.BeTrue())
	})

	// author: kuiwang@redhat.com
	g.It("Author:kuiwang-ROSA-OSD_CCS-ARO-ConnectedOnly-NonHyperShiftHOST-Medium-81538-preflight check on permission on allns mode [NONDEFAULT]", func() {
		if !olmv1util.IsFeaturegateEnabled(oc, "NewOLMPreflightPermissionChecks") {
			g.Skip("NewOLMPreflightPermissionChecks is not enable, so skip it")
		}
		exutil.SkipForSNOCluster(oc)
		var (
			caseID                   = "81538"
			ns                       = "ns-" + caseID
			sa                       = "sa" + caseID
			labelValue               = caseID
			catalogName              = "clustercatalog-" + caseID
			ceName                   = "ce-" + caseID
			clusterroleName          = ceName + "-clusterrole"
			roleName                 = ceName + "-role" + "-" + ns
			baseDir                  = exutil.FixturePath("testdata", "olm", "v1")
			clustercatalogTemplate   = filepath.Join(baseDir, "clustercatalog-withlabel.yaml")
			clusterextensionTemplate = filepath.Join(baseDir, "clusterextension-withselectorlabel.yaml")
			saTemplate               = filepath.Join(baseDir, "sa.yaml")
			bindingTemplate          = filepath.Join(baseDir, "binding-prefligth.yaml")
			clusterroleTemplate      = filepath.Join(baseDir, "prefligth-clusterrole.yaml")
			clustercatalog           = olmv1util.ClusterCatalogDescription{
				Name:       catalogName,
				Imageref:   "quay.io/olmqe/nginx-ok-index:vokv81538",
				LabelValue: labelValue,
				Template:   clustercatalogTemplate,
			}
			ce = olmv1util.ClusterExtensionDescription{
				Name:             ceName,
				PackageName:      "nginx-ok-v81538",
				Channel:          "alpha",
				Version:          ">=0.0.1",
				InstallNamespace: ns,
				SaName:           sa,
				LabelValue:       labelValue,
				Template:         clusterextensionTemplate,
			}
		)

		exutil.By("Create namespace")
		defer oc.WithoutNamespace().AsAdmin().Run("delete").Args("ns", ns, "--ignore-not-found", "--force").Execute()
		err := oc.WithoutNamespace().AsAdmin().Run("create").Args("ns", ns).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())
		o.Expect(olmv1util.Appearance(oc, exutil.Appear, "ns", ns)).To(o.BeTrue())

		exutil.By("Create clustercatalog")
		defer clustercatalog.Delete(oc)
		clustercatalog.Create(oc)

		exutil.By("create sa")
		paremeters := []string{"-n", "default", "--ignore-unknown-parameters=true", "-f", saTemplate, "-p",
			"NAME=" + sa, "NAMESPACE=" + ns}
		configFileSa, errApplySa := olmv1util.ApplyNamepsaceResourceFromTemplate(oc, ns, paremeters...)
		o.Expect(errApplySa).NotTo(o.HaveOccurred())
		defer oc.AsAdmin().WithoutNamespace().Run("delete").Args("-f", configFileSa).Execute()

		exutil.By("create clusterrole with wrong rule")
		paremeters = []string{"-n", "default", "--ignore-unknown-parameters=true", "-f", clusterroleTemplate, "-p",
			"NAME=" + clusterroleName}
		configFileCLusterroe, errApplyCLusterrole := olmv1util.ApplyClusterResourceFromTemplate(oc, paremeters...)
		o.Expect(errApplyCLusterrole).NotTo(o.HaveOccurred())
		defer oc.AsAdmin().WithoutNamespace().Run("delete").Args("-f", configFileCLusterroe).Execute()

		exutil.By("create binding")
		paremeters = []string{"-n", "default", "--ignore-unknown-parameters=true", "-f", bindingTemplate, "-p",
			"SANAME=" + sa, "NAMESPACE=" + ns, "ROLENAME=" + roleName, "CLUSTERROLESANAME=" + clusterroleName}
		configFileBinding, errApplyBinding := olmv1util.ApplyClusterResourceFromTemplate(oc, paremeters...)
		o.Expect(errApplyBinding).NotTo(o.HaveOccurred())
		defer oc.AsAdmin().WithoutNamespace().Run("delete").Args("-f", configFileBinding).Execute()

		exutil.By("check missing rule")
		defer ce.Delete(oc)
		ce.CreateWithoutCheck(oc)
		ce.CheckClusterExtensionCondition(oc, "Progressing", "message",
			`Namespace:"" Verbs:[get] NonResourceURLs:[/metrics]`, 3, 150, 0)
		ce.CheckClusterExtensionCondition(oc, "Progressing", "message",
			`Namespace:"ns-81538" APIGroups:[] Resources:[services] ResourceNames:[nginx-ok-v81538-controller-manager-metrics-service] Verbs:[delete,get,patch,update]`, 3, 150, 0)
		ce.CheckClusterExtensionCondition(oc, "Progressing", "message",
			`Namespace:"" APIGroups:[olm.operatorframework.io] Resources:[clusterextensions/finalizers] ResourceNames:[ce-81538] Verbs:[update]`, 3, 150, 0)

		exutil.By("generate rbac per missing rule and delete ce")
		jsonpath := fmt.Sprintf(`jsonpath={.status.conditions[?(@.type=="%s")].%s}`, "Progressing", "message")
		output, errGet := olmv1util.GetNoEmpty(oc, "clusterextension", ce.Name, "-o", jsonpath)
		o.Expect(errGet).NotTo(o.HaveOccurred())
		e2e.Logf("====%v====", output)

		start := "permissions to manage cluster extension:"
		end1 := "authorization evaluation error:"
		end2 := "for resolved bundle" // for current load which has no PR 1938
		filtered := olmv1util.FilterPermissions(output, start, end1, end2)
		e2e.Logf("===============================================================================")
		e2e.Logf("%v", filtered)
		e2e.Logf("===============================================================================")
		rabcDir := e2e.TestContext.OutputDir
		clusterroleFile := filepath.Join(rabcDir, fmt.Sprintf("%s.yaml", clusterroleName))
		roleFile := filepath.Join(rabcDir, fmt.Sprintf("%s.yaml", roleName))
		errGen := olmv1util.GenerateRBACFromMissingRules(filtered, ceName, rabcDir)
		o.Expect(errGen).NotTo(o.HaveOccurred())

		exutil.By("create clusterrole")
		// defer oc.AsAdmin().WithoutNamespace().Run("delete").Args("-f", clusterroleFile).Execute() // it is not needed because it is same to the above
		err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", clusterroleFile).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		exutil.By("create role")
		defer oc.AsAdmin().WithoutNamespace().Run("delete").Args("-f", roleFile).Execute()
		err = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", roleFile).Execute()
		o.Expect(err).NotTo(o.HaveOccurred())

		exutil.By("check ce again afrer applying correct rules")
		ce.CheckClusterExtensionCondition(oc, "Progressing", "reason", "Succeeded", 10, 600, 0)
	})


})
