package eventrouter

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	jsonStr = `{
		"type":"Warning",
		"reason":"CannotCreateExternalResource",
		"deploymentId":"XXXXX",
		"time":1660835529,
		"message":"cannot create EKS node group: ResourceInUseException: Cluster: test-1 is not in a valid state",
		"source":"managed/nodegroup",
		"involvedObject":{
			"apiVersion":"eks.aws.crossplane.io/v1alpha1",
			"kind":"NodeGroup",
			"name":"test-1-ng",
			"namespace": "default",
			"uid":"6365c158-8ee1-4d36-a33a-ba3cc0958ee0"
		},
		"metadata":{
			"creationTimestamp":"2022-10-27T13:10:24+02:00",
			"name":"test-1-ng.170c791ccd13d0cd",
			"namespace":"default",
			"uid":"a81f9137-9aa5-444b-95b2-7dec89c25779"
		}
	}`

	want = `"6365c158-8ee1-4d36-a33a-ba3cc0958ee0";name=test-1-ng;namespace=default;kind=NodeGroup;api-version=eks.aws.crossplane.io/v1alpha1`
)

func TestToSFV(t *testing.T) {
	var evt EventInfo
	err := json.Unmarshal([]byte(jsonStr), &evt)
	assert.Nil(t, err, "expecting nil error decoding json event")

	got, err := ToSFV(&evt)
	assert.Nil(t, err, "expecting nil error creating structured field values")
	//fmt.Println(got)
	assert.Equal(t, want, got)
}
