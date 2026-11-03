package templatesecret

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mittwald/kubernetes-secret-generator/pkg/apis/secretgenerator/v1alpha1"
)

func newTestTemplateSecret(regenerate bool) *v1alpha1.TemplateSecret {
	return &v1alpha1.TemplateSecret{
		Spec: v1alpha1.TemplateSecretSpec{
			ForceRegenerate: regenerate,
			Data: map[string]string{
				"config.env": "PASSWORD={{ .password }}",
				"other":      "USER=admin PASSWORD={{ .password }}",
			},
			Fields: []v1alpha1.TemplateField{
				{FieldName: "password", Length: "20", Encoding: "ascii"},
			},
		},
	}
}

func TestMergeTemplateDataKeepsExistingValues(t *testing.T) {
	existing := map[string][]byte{
		"config.env": []byte("PASSWORD=existing"),
		"other":      []byte("USER=admin PASSWORD=existing"),
		"unrelated":  []byte("keep-me"),
	}

	merged, err := mergeTemplateData(newTestTemplateSecret(false), existing)
	require.NoError(t, err)
	require.Equal(t, existing, merged)
}

func TestMergeTemplateDataFillsOnlyMissingKeys(t *testing.T) {
	existing := map[string][]byte{
		"config.env": []byte("PASSWORD=existing"),
		"other":      {},
	}

	merged, err := mergeTemplateData(newTestTemplateSecret(false), existing)
	require.NoError(t, err)
	require.Equal(t, "PASSWORD=existing", string(merged["config.env"]))
	require.Regexp(t, `^USER=admin PASSWORD=.{20}$`, string(merged["other"]))
}

func TestMergeTemplateDataForceRegenerate(t *testing.T) {
	existing := map[string][]byte{
		"config.env": []byte("PASSWORD=existing"),
		"other":      []byte("USER=admin PASSWORD=existing"),
		"unrelated":  []byte("keep-me"),
	}

	merged, err := mergeTemplateData(newTestTemplateSecret(true), existing)
	require.NoError(t, err)
	require.Regexp(t, `^PASSWORD=.{20}$`, string(merged["config.env"]))
	require.NotEqual(t, "PASSWORD=existing", string(merged["config.env"]))
	require.Regexp(t, `^USER=admin PASSWORD=.{20}$`, string(merged["other"]))
	require.Equal(t, "keep-me", string(merged["unrelated"]))
}
