package plugin

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// VolumeSnapshotter is a plugin for containing state for the longhorn volume
type VolumeSnapshotter struct {
	Log logrus.FieldLogger
}

// NewVolumeSnapshotter instantiates a NewVolumeSnapshotter.
func NewVolumeSnapshotter(log logrus.FieldLogger) *VolumeSnapshotter {
	return &VolumeSnapshotter{Log: log}
}

// Init prepares the VolumeSnapshotter for usage using the provided map of
// configuration key-value pairs. It returns an error if the VolumeSnapshotter
// cannot be initialized from the provided config. Note that after v0.10.0, this will happen multiple times.
func (vs *VolumeSnapshotter) Init(config map[string]string) error {
	vs.Log.Infof("Init called", config)
	return nil
}

// CreateVolumeFromSnapshot creates a new volume in the specified
// availability zone, initialized from the provided snapshot,
// and with the specified type and IOPS (if using provisioned IOPS).
func (vs *VolumeSnapshotter) CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error) {
	vs.Log.Infof("CreateVolumeFromSnapshot called", snapshotID, volumeType, volumeAZ, *iops)
	return "", nil
}

// GetVolumeInfo returns the type and IOPS (if using provisioned IOPS) for
// the specified volume in the given availability zone.
func (vs *VolumeSnapshotter) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	vs.Log.Infof("GetVolumeInfo called", volumeID, volumeAZ)
	return "", nil, nil
}

// IsVolumeReady Check if the volume is ready.
func (vs *VolumeSnapshotter) IsVolumeReady(volumeID, volumeAZ string) (ready bool, err error) {
	vs.Log.Infof("IsVolumeReady called", volumeID, volumeAZ)
	return true, nil
}

// CreateSnapshot creates a snapshot of the specified volume, and applies any provided
// set of tags to the snapshot.
func (vs *VolumeSnapshotter) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
	vs.Log.Infof("CreateSnapshot called", volumeID, volumeAZ, tags)
	return "", nil
}

// DeleteSnapshot deletes the specified volume snapshot.
func (vs *VolumeSnapshotter) DeleteSnapshot(snapshotID string) error {
	vs.Log.Infof("DeleteSnapshot called", snapshotID)
	return nil
}

// GetVolumeID returns the specific identifier for the PersistentVolume.
func (vs *VolumeSnapshotter) GetVolumeID(unstructuredPV runtime.Unstructured) (string, error) {
	vs.Log.Infof("GetVolumeID called", unstructuredPV)

	pv := new(corev1.PersistentVolume)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPV.UnstructuredContent(), pv); err != nil {
		return "", errors.WithStack(err)
	}

	if pv.Spec.CSI == nil {
		return "", fmt.Errorf("unable to retrieve CSI Spec from pv %+v", pv)
	}
	if pv.Spec.CSI.VolumeHandle == "" {
		return "", fmt.Errorf("unable to retrieve Volume handle from pv %+v", pv)
	}
	return pv.Spec.CSI.VolumeHandle, nil
}

// SetVolumeID sets the specific identifier for the PersistentVolume.
func (vs *VolumeSnapshotter) SetVolumeID(unstructuredPV runtime.Unstructured, volumeID string) (runtime.Unstructured, error) {
	vs.Log.Infof("SetVolumeID called", unstructuredPV, volumeID)

	pv := new(corev1.PersistentVolume)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPV.UnstructuredContent(), pv); err != nil {
		return nil, errors.WithStack(err)
	}

	if pv.Spec.CSI == nil {
		return nil, fmt.Errorf("spec.CSI not found from pv %+v", pv)
	}

	pv.Spec.CSI.VolumeHandle = volumeID

	res, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pv)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &unstructured.Unstructured{Object: res}, nil
}
