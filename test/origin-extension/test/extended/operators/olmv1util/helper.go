package olmv1util

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	o "github.com/onsi/gomega"
	exutil "github.com/openshift/openshift-tests-private/test/extended/util"
)

// it is used to get OLMv1 resource's field.
// if ns is needed, please add "-n" in parameters
// it take 3s and 150s as default value for wait.Poll. if it is not ok later, could change it.
func Get(oc *exutil.CLI, parameters ...string) (string, error) {
	return exutil.GetFieldWithJsonpath(oc, 3*time.Second, 150*time.Second, exutil.Immediately,
		exutil.AllowEmpty, exutil.AsAdmin, exutil.WithoutNamespace, parameters...)
}

// it is same to Get except that it does not alllow to return empty string.
func GetNoEmpty(oc *exutil.CLI, parameters ...string) (string, error) {
	return exutil.GetFieldWithJsonpath(oc, 3*time.Second, 150*time.Second, exutil.Immediately,
		exutil.NotAllowEmpty, exutil.AsAdmin, exutil.WithoutNamespace, parameters...)
}

func Cleanup(oc *exutil.CLI, parameters ...string) {
	exutil.CleanupResource(oc, 4*time.Second, 160*time.Second,
		exutil.AsAdmin, exutil.WithoutNamespace, parameters...)
}

func resourceFromTemplate(oc *exutil.CLI, create bool, returnError bool, namespace string, parameters ...string) (string, error) {
	var configFile string
	errWait := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 15*time.Second,
		false, func(ctx context.Context) (bool, error) {
			fileName := exutil.GetRandomString() + "config.json"
			stdout, _, err := oc.AsAdmin().Run("process").Args(parameters...).OutputsToFiles(fileName)
			if err != nil {
				e2e.Logf("the err:%v, and try next round", err)
				return false, nil
			}

			configFile = stdout
			return true, nil
		})
	if returnError && errWait != nil {
		e2e.Logf("fail to process %v", parameters)
		return "", errWait
	}
	exutil.AssertWaitPollNoErr(errWait, fmt.Sprintf("fail to process %v", parameters))

	e2e.Logf("the file of resource is %s", configFile)

	var resourceErr error
	if create {
		if namespace != "" {
			resourceErr = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", configFile, "-n", namespace).Execute()
		} else {
			resourceErr = oc.AsAdmin().WithoutNamespace().Run("create").Args("-f", configFile).Execute()
		}
	} else {
		if namespace != "" {
			resourceErr = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", configFile, "-n", namespace).Execute()
		} else {
			resourceErr = oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", configFile).Execute()
		}
	}
	if returnError && resourceErr != nil {
		e2e.Logf("fail to create/apply resource %v", resourceErr)
		return "", resourceErr
	}
	exutil.AssertWaitPollNoErr(resourceErr, fmt.Sprintf("fail to create/apply resource %v", resourceErr))
	return configFile, nil
}

func ApplyClusterResourceFromTemplate(oc *exutil.CLI, parameters ...string) (string, error) {
	return resourceFromTemplate(oc, false, true, "", parameters...)
}
func ApplyNamepsaceResourceFromTemplate(oc *exutil.CLI, namespace string, parameters ...string) (string, error) {
	return resourceFromTemplate(oc, false, true, namespace, parameters...)
}

func FilterPermissions(msg, startMarker, end1Marker, end2Marker string) string {
	scanner := bufio.NewScanner(strings.NewReader(msg))
	var sb strings.Builder
	inSection := false

	for scanner.Scan() {
		line := scanner.Text()
		// Start when we see the start marker
		if !inSection && strings.Contains(line, startMarker) {
			inSection = true
			continue
		}
		if inSection {
			// Check if this line contains the end marker
			if idx := strings.Index(line, end1Marker); idx != -1 {
				// Include content before the marker and then stop
				part := strings.TrimSpace(line[:idx])
				sb.WriteString(part)
				sb.WriteString("\n")
				break
			}
			if idx := strings.Index(line, end2Marker); idx != -1 {
				// Include content before the marker and then stop
				part := strings.TrimSpace(line[:idx])
				sb.WriteString(part)
				sb.WriteString("\n")
				break
			}
			// Otherwise, include the full line
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func Appearance(oc *exutil.CLI, appear bool, parameters ...string) bool {
	return exutil.CheckAppearance(oc, 4*time.Second, 200*time.Second, exutil.NotImmediately,
		exutil.AsAdmin, exutil.WithoutNamespace, appear, parameters...)
}

// IsFeaturegateEnabled checks if a cluster enable feature gate
func IsFeaturegateEnabled(oc *exutil.CLI, featuregate string) bool {
	featureGate, err := oc.AdminConfigClient().ConfigV1().FeatureGates().Get(context.Background(), "cluster", metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false
		}
		o.Expect(err).NotTo(o.HaveOccurred(), "could not retrieve feature-gate: %v", err)
	}

	isEnabled := false
	for _, featureGate := range featureGate.Status.FeatureGates {
		for _, enabled := range featureGate.Enabled {
			if string(enabled.Name) == featuregate {
				isEnabled = true
				break
			}
		}
		if isEnabled {
			break
		}
	}
	return isEnabled
}
