package eventrouter

import (
	"time"

	"github.com/ucarion/sfv"
)

type InvolvedObject struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace"`
	UID        string `json:"uid"`
}

type Metadata struct {
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	UID               string    `json:"uid"`
}

type EventInfo struct {
	Type           string         `json:"type"`
	Reason         string         `json:"reason"`
	DeploymentId   string         `json:"deploymentId"`
	Time           int64          `json:"time"`
	Message        string         `json:"message"`
	Source         string         `json:"source"`
	InvolvedObject InvolvedObject `json:"involvedObject"`
	Metadata       Metadata       `json:"metadata"`
}

/*
func ToSubject(evt *EventInfo) string {
	var sb strings.Builder
	sb.WriteString(evt.InvolvedObject.UID)
	sb.WriteRune(';')
	sb.WriteString(fmt.Sprintf("name=%s", evt.InvolvedObject.Name))
	sb.WriteRune(';')
	sb.WriteString(fmt.Sprintf("namespace=%s", evt.InvolvedObject.Namespace))
	sb.WriteRune(';')
	sb.WriteString(fmt.Sprintf("apiversion=%s", evt.InvolvedObject.APIVersion))
	sb.WriteRune(';')
	sb.WriteString(fmt.Sprintf("kind=%s", evt.InvolvedObject.Kind))
	return sb.String()
}
*/

func ToSFV(evt *EventInfo) (string, error) {
	item := sfv.Item{
		BareItem: sfv.BareItem{
			Type:   sfv.BareItemTypeString,
			String: evt.InvolvedObject.UID,
		},

		Params: sfv.Params{
			Keys: []string{"name", "namespace", "kind", "api-version"},
			Map: map[string]sfv.BareItem{
				"name": {
					Type:  sfv.BareItemTypeToken,
					Token: evt.InvolvedObject.Name,
				},
				"namespace": {
					Type:  sfv.BareItemTypeToken,
					Token: evt.InvolvedObject.Namespace,
				},
				"kind": {
					Type:  sfv.BareItemTypeToken,
					Token: evt.InvolvedObject.Kind,
				},
				"api-version": {
					Type:  sfv.BareItemTypeToken,
					Token: evt.InvolvedObject.APIVersion,
				},
			},
		},
	}

	return sfv.Marshal(item)
}
