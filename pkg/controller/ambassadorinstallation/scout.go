// From ambassador/internal/pkg/edgectl/scout.go with some modifications & simplifications.

package ambassadorinstallation

import (
	"context"

	"k8s.io/apimachinery/pkg/types"

	"github.com/datawire/ambassador-operator/version"
	"github.com/datawire/ambassador/pkg/metriton"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// The Scout structure maintains an index, which is the count of calls
// to Report, incrementing each call.  This provides a sequence of actions
// to make it easier to search through the reports.
// The Reporter is simply the Metriton reporting object.
type Scout struct {
	index    int
	Reporter *metriton.Reporter
}

// Metadata is simply a key and an untyped value, instances passed in as parameters
// to (s *Scout) Report
type ScoutMeta struct {
	Key   string
	Value interface{}
}

// Create a new Scout object, with a parameter stating what the Scout instance
// will be reporting on.  The Ambassador Operator may be installing, updating,
// or deleting the Ambassador installation.
func NewScout(mode string, installID types.UID) (s *Scout) {
	return &Scout{
		index: 0,
		Reporter: &metriton.Reporter{
			Application: "ambassador-operator",
			Version:     version.Version,
			GetInstallID: func(r *metriton.Reporter) (string, error) {
				return string(installID), nil
			},
			// Fixed (growing) metadata passed with every report
			BaseMetadata: map[string]interface{}{
				"mode":     mode,
				"trace_id": uuid.New().String(),
			},
		},
	}
}

// Reporting out: Sends a report to Metriton which will create a new entry in the
// Metriton database in the product_event table.
func (s *Scout) Report(action string, meta ...ScoutMeta) error {
	// Construct the report's metadata. Include the fixed (growing) set of
	// metadata in the Scout structure and the pairs passed as arguments to this
	// call. Also include and increment the index, which can be used to
	// determine the correct order of reported events for this installation
	// attempt (correlated by the trace_id set at the start).
	s.index++
	metadata := map[string]interface{}{
		"action": action,
		"index":  s.index,
	}
	for _, metaItem := range meta {
		metadata[metaItem.Key] = metaItem.Value
	}

	// TODO: @Alvaro, please check--is this the context we want to pass through
	// TODO to Metriton?
	_, err := s.Reporter.Report(context.TODO(), metadata)
	if err != nil {
		return errors.Wrap(err, "scout report")
	}

	// TODO: Do something useful (alert the user if there's an available
	// upgrade?) with the response (discarded as "_" above)?

	return nil
}
