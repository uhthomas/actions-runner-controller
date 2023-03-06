package tests

import (
	"path/filepath"
	"strings"
	"testing"

	v1alpha1 "github.com/actions/actions-runner-controller/apis/actions.github.com/v1alpha1"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestTemplateRenderedGitHubSecretWithGitHubToken(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/githubsecret.yaml"})

	var githubSecret corev1.Secret
	helm.UnmarshalK8SYaml(t, output, &githubSecret)

	assert.Equal(t, namespaceName, githubSecret.Namespace)
	assert.Equal(t, "test-runners-gha-runner-scale-set-github-secret", githubSecret.Name)
	assert.Equal(t, "gh_token12345", string(githubSecret.Data["github_token"]))
	assert.Equal(t, "actions.github.com/secret-protection", githubSecret.Finalizers[0])
}

func TestTemplateRenderedGitHubSecretWithGitHubApp(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                               "https://github.com/actions",
			"githubConfigSecret.github_app_id":              "10",
			"githubConfigSecret.github_app_installation_id": "100",
			"githubConfigSecret.github_app_private_key":     "private_key",
			"controllerServiceAccount.name":                 "arc",
			"controllerServiceAccount.namespace":            "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/githubsecret.yaml"})

	var githubSecret corev1.Secret
	helm.UnmarshalK8SYaml(t, output, &githubSecret)

	assert.Equal(t, namespaceName, githubSecret.Namespace)
	assert.Equal(t, "10", string(githubSecret.Data["github_app_id"]))
	assert.Equal(t, "100", string(githubSecret.Data["github_app_installation_id"]))
	assert.Equal(t, "private_key", string(githubSecret.Data["github_app_private_key"]))
}

func TestTemplateRenderedGitHubSecretErrorWithMissingAuthInput(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_app_id":   "",
			"githubConfigSecret.github_token":    "",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/githubsecret.yaml"})
	require.Error(t, err)

	assert.ErrorContains(t, err, "provide .Values.githubConfigSecret.github_token or .Values.githubConfigSecret.github_app_id")
}

func TestTemplateRenderedGitHubSecretErrorWithMissingAppInput(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_app_id":   "10",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/githubsecret.yaml"})
	require.Error(t, err)

	assert.ErrorContains(t, err, "provide .Values.githubConfigSecret.github_app_installation_id and .Values.githubConfigSecret.github_app_private_key")
}

func TestTemplateNotRenderedGitHubSecretWithPredefinedSecret(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret":                 "pre-defined-secret",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/githubsecret.yaml"})
	assert.ErrorContains(t, err, "could not find template templates/githubsecret.yaml in chart", "secret should not be rendered since a pre-defined secret is provided")
}

func TestTemplateRenderedSetServiceAccountToNoPermission(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/no_permission_serviceaccount.yaml"})
	var serviceAccount corev1.ServiceAccount
	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	assert.Equal(t, namespaceName, serviceAccount.Namespace)
	assert.Equal(t, "test-runners-gha-runner-scale-set-no-permission-service-account", serviceAccount.Name)

	output = helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})
	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, "test-runners-gha-runner-scale-set-no-permission-service-account", ars.Spec.Template.Spec.ServiceAccountName)
}

func TestTemplateRenderedSetServiceAccountToKubeMode(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"containerMode.type":                 "kubernetes",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/kube_mode_serviceaccount.yaml"})
	var serviceAccount corev1.ServiceAccount
	helm.UnmarshalK8SYaml(t, output, &serviceAccount)

	assert.Equal(t, namespaceName, serviceAccount.Namespace)
	assert.Equal(t, "test-runners-gha-runner-scale-set-kube-mode-service-account", serviceAccount.Name)

	output = helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/kube_mode_role.yaml"})
	var role rbacv1.Role
	helm.UnmarshalK8SYaml(t, output, &role)

	assert.Equal(t, namespaceName, role.Namespace)
	assert.Equal(t, "test-runners-gha-runner-scale-set-kube-mode-role", role.Name)
	assert.Len(t, role.Rules, 5, "kube mode role should have 5 rules")
	assert.Equal(t, "pods", role.Rules[0].Resources[0])
	assert.Equal(t, "pods/exec", role.Rules[1].Resources[0])
	assert.Equal(t, "pods/log", role.Rules[2].Resources[0])
	assert.Equal(t, "jobs", role.Rules[3].Resources[0])
	assert.Equal(t, "secrets", role.Rules[4].Resources[0])

	output = helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/kube_mode_role_binding.yaml"})
	var roleBinding rbacv1.RoleBinding
	helm.UnmarshalK8SYaml(t, output, &roleBinding)

	assert.Equal(t, namespaceName, roleBinding.Namespace)
	assert.Equal(t, "test-runners-gha-runner-scale-set-kube-mode-role", roleBinding.Name)
	assert.Len(t, roleBinding.Subjects, 1)
	assert.Equal(t, "test-runners-gha-runner-scale-set-kube-mode-service-account", roleBinding.Subjects[0].Name)
	assert.Equal(t, namespaceName, roleBinding.Subjects[0].Namespace)
	assert.Equal(t, "test-runners-gha-runner-scale-set-kube-mode-role", roleBinding.RoleRef.Name)
	assert.Equal(t, "Role", roleBinding.RoleRef.Kind)

	output = helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})
	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, "test-runners-gha-runner-scale-set-kube-mode-service-account", ars.Spec.Template.Spec.ServiceAccountName)
}

func TestTemplateRenderedUserProvideSetServiceAccount(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"template.spec.serviceAccountName":   "test-service-account",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/no_permission_serviceaccount.yaml"})
	assert.ErrorContains(t, err, "could not find template templates/no_permission_serviceaccount.yaml in chart", "no permission service account should not be rendered")

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})
	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, "test-service-account", ars.Spec.Template.Spec.ServiceAccountName)
}

func TestTemplateRenderedAutoScalingRunnerSet(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, namespaceName, ars.Namespace)
	assert.Equal(t, "test-runners", ars.Name)

	assert.Equal(t, "gha-runner-scale-set", ars.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "test-runners", ars.Labels["app.kubernetes.io/instance"])
	assert.Equal(t, "https://github.com/actions", ars.Spec.GitHubConfigUrl)
	assert.Equal(t, "test-runners-gha-runner-scale-set-github-secret", ars.Spec.GitHubConfigSecret)

	assert.Empty(t, ars.Spec.RunnerGroup, "RunnerGroup should be empty")

	assert.Nil(t, ars.Spec.MinRunners, "MinRunners should be nil")
	assert.Nil(t, ars.Spec.MaxRunners, "MaxRunners should be nil")
	assert.Nil(t, ars.Spec.Proxy, "Proxy should be nil")
	assert.Nil(t, ars.Spec.GitHubServerTLS, "GitHubServerTLS should be nil")

	assert.NotNil(t, ars.Spec.Template.Spec, "Template.Spec should not be nil")

	assert.Len(t, ars.Spec.Template.Spec.Containers, 1, "Template.Spec should have 1 container")
	assert.Equal(t, "runner", ars.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "ghcr.io/actions/actions-runner:latest", ars.Spec.Template.Spec.Containers[0].Image)
}

func TestTemplateRenderedAutoScalingRunnerSet_ProvideMetadata(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                     "https://github.com/actions",
			"githubConfigSecret.github_token":     "gh_token12345",
			"template.metadata.labels.test1":      "test1",
			"template.metadata.labels.test2":      "test2",
			"template.metadata.annotations.test3": "test3",
			"template.metadata.annotations.test4": "test4",
			"controllerServiceAccount.name":       "arc",
			"controllerServiceAccount.namespace":  "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, namespaceName, ars.Namespace)
	assert.Equal(t, "test-runners", ars.Name)

	assert.NotNil(t, ars.Spec.Template.Labels, "Template.Spec.Labels should not be nil")
	assert.Equal(t, "test1", ars.Spec.Template.Labels["test1"], "Template.Spec.Labels should have test1")
	assert.Equal(t, "test2", ars.Spec.Template.Labels["test2"], "Template.Spec.Labels should have test2")

	assert.NotNil(t, ars.Spec.Template.Annotations, "Template.Spec.Annotations should not be nil")
	assert.Equal(t, "test3", ars.Spec.Template.Annotations["test3"], "Template.Spec.Annotations should have test3")
	assert.Equal(t, "test4", ars.Spec.Template.Annotations["test4"], "Template.Spec.Annotations should have test4")

	assert.NotNil(t, ars.Spec.Template.Spec, "Template.Spec should not be nil")

	assert.Len(t, ars.Spec.Template.Spec.Containers, 1, "Template.Spec should have 1 container")
	assert.Equal(t, "runner", ars.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "ghcr.io/actions/actions-runner:latest", ars.Spec.Template.Spec.Containers[0].Image)
}

func TestTemplateRenderedAutoScalingRunnerSet_MaxRunnersValidationError(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"maxRunners":                         "-1",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})
	require.Error(t, err)

	assert.ErrorContains(t, err, "maxRunners has to be greater or equal to 0")
}

func TestTemplateRenderedAutoScalingRunnerSet_MinRunnersValidationError(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"maxRunners":                         "1",
			"minRunners":                         "-1",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})
	require.Error(t, err)

	assert.ErrorContains(t, err, "minRunners has to be greater or equal to 0")
}

func TestTemplateRenderedAutoScalingRunnerSet_MinMaxRunnersValidationError(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"maxRunners":                         "0",
			"minRunners":                         "1",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})
	require.Error(t, err)

	assert.ErrorContains(t, err, "maxRunners has to be greater or equal to minRunners")
}

func TestTemplateRenderedAutoScalingRunnerSet_MinMaxRunnersValidationSameValue(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"maxRunners":                         "0",
			"minRunners":                         "0",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, 0, *ars.Spec.MinRunners, "MinRunners should be 0")
	assert.Equal(t, 0, *ars.Spec.MaxRunners, "MaxRunners should be 0")
}

func TestTemplateRenderedAutoScalingRunnerSet_MinMaxRunnersValidation_OnlyMin(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"minRunners":                         "5",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, 5, *ars.Spec.MinRunners, "MinRunners should be 5")
	assert.Nil(t, ars.Spec.MaxRunners, "MaxRunners should be nil")
}

func TestTemplateRenderedAutoScalingRunnerSet_MinMaxRunnersValidation_OnlyMax(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"maxRunners":                         "5",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, 5, *ars.Spec.MaxRunners, "MaxRunners should be 5")
	assert.Nil(t, ars.Spec.MinRunners, "MinRunners should be nil")
}

func TestTemplateRenderedAutoScalingRunnerSet_MinMaxRunners_FromValuesFile(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	testValuesPath, err := filepath.Abs("../tests/values.yaml")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		ValuesFiles:    []string{testValuesPath},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, 5, *ars.Spec.MinRunners, "MinRunners should be 5")
	assert.Equal(t, 10, *ars.Spec.MaxRunners, "MaxRunners should be 10")
}

func TestTemplateRenderedAutoScalingRunnerSet_EnableDinD(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"containerMode.type":                 "dind",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, namespaceName, ars.Namespace)
	assert.Equal(t, "test-runners", ars.Name)

	assert.Equal(t, "gha-runner-scale-set", ars.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "test-runners", ars.Labels["app.kubernetes.io/instance"])
	assert.Equal(t, "https://github.com/actions", ars.Spec.GitHubConfigUrl)
	assert.Equal(t, "test-runners-gha-runner-scale-set-github-secret", ars.Spec.GitHubConfigSecret)

	assert.Empty(t, ars.Spec.RunnerGroup, "RunnerGroup should be empty")

	assert.Nil(t, ars.Spec.MinRunners, "MinRunners should be nil")
	assert.Nil(t, ars.Spec.MaxRunners, "MaxRunners should be nil")
	assert.Nil(t, ars.Spec.Proxy, "Proxy should be nil")
	assert.Nil(t, ars.Spec.GitHubServerTLS, "GitHubServerTLS should be nil")

	assert.NotNil(t, ars.Spec.Template.Spec, "Template.Spec should not be nil")

	assert.Len(t, ars.Spec.Template.Spec.InitContainers, 1, "Template.Spec should have 1 init container")
	assert.Equal(t, "init-dind-externals", ars.Spec.Template.Spec.InitContainers[0].Name)
	assert.Equal(t, "ghcr.io/actions/actions-runner:latest", ars.Spec.Template.Spec.InitContainers[0].Image)
	assert.Equal(t, "cp", ars.Spec.Template.Spec.InitContainers[0].Command[0])
	assert.Equal(t, "-r -v /actions-runner/externals/. /actions-runner/tmpDir/", strings.Join(ars.Spec.Template.Spec.InitContainers[0].Args, " "))

	assert.Len(t, ars.Spec.Template.Spec.Containers, 2, "Template.Spec should have 2 container")
	assert.Equal(t, "runner", ars.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "ghcr.io/actions/actions-runner:latest", ars.Spec.Template.Spec.Containers[0].Image)
	assert.Len(t, ars.Spec.Template.Spec.Containers[0].Env, 4, "The runner container should have 4 env vars, DOCKER_HOST, DOCKER_TLS_VERIFY, DOCKER_CERT_PATH and RUNNER_WAIT_FOR_DOCKER_IN_SECONDS")
	assert.Equal(t, "DOCKER_HOST", ars.Spec.Template.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "tcp://localhost:2376", ars.Spec.Template.Spec.Containers[0].Env[0].Value)
	assert.Equal(t, "DOCKER_TLS_VERIFY", ars.Spec.Template.Spec.Containers[0].Env[1].Name)
	assert.Equal(t, "1", ars.Spec.Template.Spec.Containers[0].Env[1].Value)
	assert.Equal(t, "DOCKER_CERT_PATH", ars.Spec.Template.Spec.Containers[0].Env[2].Name)
	assert.Equal(t, "/certs/client", ars.Spec.Template.Spec.Containers[0].Env[2].Value)
	assert.Equal(t, "RUNNER_WAIT_FOR_DOCKER_IN_SECONDS", ars.Spec.Template.Spec.Containers[0].Env[3].Name)
	assert.Equal(t, "120", ars.Spec.Template.Spec.Containers[0].Env[3].Value)

	assert.Len(t, ars.Spec.Template.Spec.Containers[0].VolumeMounts, 2, "The runner container should have 2 volume mounts, dind-cert and work")
	assert.Equal(t, "work", ars.Spec.Template.Spec.Containers[0].VolumeMounts[0].Name)
	assert.Equal(t, "/actions-runner/_work", ars.Spec.Template.Spec.Containers[0].VolumeMounts[0].MountPath)
	assert.False(t, ars.Spec.Template.Spec.Containers[0].VolumeMounts[0].ReadOnly)

	assert.Equal(t, "dind-cert", ars.Spec.Template.Spec.Containers[0].VolumeMounts[1].Name)
	assert.Equal(t, "/certs/client", ars.Spec.Template.Spec.Containers[0].VolumeMounts[1].MountPath)
	assert.True(t, ars.Spec.Template.Spec.Containers[0].VolumeMounts[1].ReadOnly)

	assert.Equal(t, "dind", ars.Spec.Template.Spec.Containers[1].Name)
	assert.Equal(t, "docker:dind", ars.Spec.Template.Spec.Containers[1].Image)
	assert.True(t, *ars.Spec.Template.Spec.Containers[1].SecurityContext.Privileged)
	assert.Len(t, ars.Spec.Template.Spec.Containers[1].VolumeMounts, 3, "The dind container should have 3 volume mounts, dind-cert, work and externals")
	assert.Equal(t, "work", ars.Spec.Template.Spec.Containers[1].VolumeMounts[0].Name)
	assert.Equal(t, "/actions-runner/_work", ars.Spec.Template.Spec.Containers[1].VolumeMounts[0].MountPath)

	assert.Equal(t, "dind-cert", ars.Spec.Template.Spec.Containers[1].VolumeMounts[1].Name)
	assert.Equal(t, "/certs/client", ars.Spec.Template.Spec.Containers[1].VolumeMounts[1].MountPath)

	assert.Equal(t, "dind-externals", ars.Spec.Template.Spec.Containers[1].VolumeMounts[2].Name)
	assert.Equal(t, "/actions-runner/externals", ars.Spec.Template.Spec.Containers[1].VolumeMounts[2].MountPath)
}

func TestTemplateRenderedAutoScalingRunnerSet_EnableKubernetesMode(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"containerMode.type":                 "kubernetes",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, namespaceName, ars.Namespace)
	assert.Equal(t, "test-runners", ars.Name)

	assert.Equal(t, "gha-runner-scale-set", ars.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "test-runners", ars.Labels["app.kubernetes.io/instance"])
	assert.Equal(t, "https://github.com/actions", ars.Spec.GitHubConfigUrl)
	assert.Equal(t, "test-runners-gha-runner-scale-set-github-secret", ars.Spec.GitHubConfigSecret)

	assert.Empty(t, ars.Spec.RunnerGroup, "RunnerGroup should be empty")
	assert.Nil(t, ars.Spec.MinRunners, "MinRunners should be nil")
	assert.Nil(t, ars.Spec.MaxRunners, "MaxRunners should be nil")
	assert.Nil(t, ars.Spec.Proxy, "Proxy should be nil")
	assert.Nil(t, ars.Spec.GitHubServerTLS, "GitHubServerTLS should be nil")

	assert.NotNil(t, ars.Spec.Template.Spec, "Template.Spec should not be nil")

	assert.Len(t, ars.Spec.Template.Spec.Containers, 1, "Template.Spec should have 1 container")
	assert.Equal(t, "runner", ars.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "ghcr.io/actions/actions-runner:latest", ars.Spec.Template.Spec.Containers[0].Image)

	assert.Equal(t, "ACTIONS_RUNNER_CONTAINER_HOOKS", ars.Spec.Template.Spec.Containers[0].Env[0].Name)
	assert.Equal(t, "/actions-runner/k8s/index.js", ars.Spec.Template.Spec.Containers[0].Env[0].Value)
	assert.Equal(t, "ACTIONS_RUNNER_POD_NAME", ars.Spec.Template.Spec.Containers[0].Env[1].Name)
	assert.Equal(t, "ACTIONS_RUNNER_REQUIRE_JOB_CONTAINER", ars.Spec.Template.Spec.Containers[0].Env[2].Name)
	assert.Equal(t, "true", ars.Spec.Template.Spec.Containers[0].Env[2].Value)

	assert.Len(t, ars.Spec.Template.Spec.Volumes, 1, "Template.Spec should have 1 volume")
	assert.Equal(t, "work", ars.Spec.Template.Spec.Volumes[0].Name)
	assert.NotNil(t, ars.Spec.Template.Spec.Volumes[0].Ephemeral, "Template.Spec should have 1 ephemeral volume")
}

func TestTemplateRenderedAutoScalingRunnerSet_UsePredefinedSecret(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret":                 "pre-defined-secrets",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	assert.Equal(t, namespaceName, ars.Namespace)
	assert.Equal(t, "test-runners", ars.Name)

	assert.Equal(t, "gha-runner-scale-set", ars.Labels["app.kubernetes.io/name"])
	assert.Equal(t, "test-runners", ars.Labels["app.kubernetes.io/instance"])
	assert.Equal(t, "https://github.com/actions", ars.Spec.GitHubConfigUrl)
	assert.Equal(t, "pre-defined-secrets", ars.Spec.GitHubConfigSecret)
}

func TestTemplateRenderedAutoScalingRunnerSet_ErrorOnEmptyPredefinedSecret(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret":                 "",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})
	require.Error(t, err)

	assert.ErrorContains(t, err, "Values.githubConfigSecret is required for setting auth with GitHub server")
}

func TestTemplateRenderedWithProxy(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret":                 "pre-defined-secrets",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
			"proxy.http.url":                     "http://proxy.example.com",
			"proxy.http.credentialSecretRef":     "http-secret",
			"proxy.https.url":                    "https://proxy.example.com",
			"proxy.https.credentialSecretRef":    "https-secret",
			"proxy.noProxy":                      "{example.com,example.org}",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/autoscalingrunnerset.yaml"})

	var ars v1alpha1.AutoscalingRunnerSet
	helm.UnmarshalK8SYaml(t, output, &ars)

	require.NotNil(t, ars.Spec.Proxy)
	require.NotNil(t, ars.Spec.Proxy.HTTP)
	assert.Equal(t, "http://proxy.example.com", ars.Spec.Proxy.HTTP.Url)
	assert.Equal(t, "http-secret", ars.Spec.Proxy.HTTP.CredentialSecretRef)

	require.NotNil(t, ars.Spec.Proxy.HTTPS)
	assert.Equal(t, "https://proxy.example.com", ars.Spec.Proxy.HTTPS.Url)
	assert.Equal(t, "https-secret", ars.Spec.Proxy.HTTPS.CredentialSecretRef)

	require.NotNil(t, ars.Spec.Proxy.NoProxy)
	require.Len(t, ars.Spec.Proxy.NoProxy, 2)
	assert.Contains(t, ars.Spec.Proxy.NoProxy, "example.com")
	assert.Contains(t, ars.Spec.Proxy.NoProxy, "example.org")
}

func TestTemplateNamingConstraints(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	setValues := map[string]string{
		"githubConfigUrl":                    "https://github.com/actions",
		"githubConfigSecret":                 "",
		"controllerServiceAccount.name":      "arc",
		"controllerServiceAccount.namespace": "arc-system",
	}

	tt := map[string]struct {
		releaseName   string
		namespaceName string
		expectedError string
	}{
		"Name too long": {
			releaseName:   strings.Repeat("a", 46),
			namespaceName: "test-" + strings.ToLower(random.UniqueId()),
			expectedError: "Name must have up to 45 characters",
		},
		"Namespace too long": {
			releaseName:   "test-" + strings.ToLower(random.UniqueId()),
			namespaceName: strings.Repeat("a", 64),
			expectedError: "Namespace must have up to 63 characters",
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			options := &helm.Options{
				SetValues:      setValues,
				KubectlOptions: k8s.NewKubectlOptions("", "", tc.namespaceName),
			}
			_, err = helm.RenderTemplateE(t, options, helmChartPath, tc.releaseName, []string{"templates/autoscalingrunnerset.yaml"})
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.expectedError)
		})
	}
}

func TestTemplate_CreateManagerRole(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/manager_role.yaml"})

	var managerRole rbacv1.Role
	helm.UnmarshalK8SYaml(t, output, &managerRole)

	assert.Equal(t, namespaceName, managerRole.Namespace, "namespace should match the namespace of the Helm release")
	assert.Equal(t, "test-runners-gha-runner-scale-set-manager-role", managerRole.Name)
	assert.Equal(t, 6, len(managerRole.Rules))
}

func TestTemplate_CreateManagerRoleBinding(t *testing.T) {
	t.Parallel()

	// Path to the helm chart we will test
	helmChartPath, err := filepath.Abs("../../gha-runner-scale-set")
	require.NoError(t, err)

	releaseName := "test-runners"
	namespaceName := "test-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues: map[string]string{
			"githubConfigUrl":                    "https://github.com/actions",
			"githubConfigSecret.github_token":    "gh_token12345",
			"controllerServiceAccount.name":      "arc",
			"controllerServiceAccount.namespace": "arc-system",
		},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/manager_role_binding.yaml"})

	var managerRoleBinding rbacv1.RoleBinding
	helm.UnmarshalK8SYaml(t, output, &managerRoleBinding)

	assert.Equal(t, namespaceName, managerRoleBinding.Namespace, "namespace should match the namespace of the Helm release")
	assert.Equal(t, "test-runners-gha-runner-scale-set-manager-role-binding", managerRoleBinding.Name)
	assert.Equal(t, "test-runners-gha-runner-scale-set-manager-role", managerRoleBinding.RoleRef.Name)
	assert.Equal(t, "arc", managerRoleBinding.Subjects[0].Name)
	assert.Equal(t, "arc-system", managerRoleBinding.Subjects[0].Namespace)
}
