package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	backendServiceURL = "http://longhorn-backend.longhorn-system.svc:9500"
)

// VolumeSnapshotter is a plugin for containing state for the longhorn volume
type VolumeSnapshotter struct {
	Log     logrus.FieldLogger
	kClient *kubernetes.Clientset
}

// NewVolumeSnapshotter instantiates a NewVolumeSnapshotter.
func NewVolumeSnapshotter(log logrus.FieldLogger) *VolumeSnapshotter {
	return &VolumeSnapshotter{Log: log}
}

// Init prepares the VolumeSnapshotter for usage using the provided map of
// configuration key-value pairs. It returns an error if the VolumeSnapshotter
// cannot be initialized from the provided config. Note that after v0.10.0, this will happen multiple times.
func (vs *VolumeSnapshotter) Init(config map[string]string) error {
	clusterConfig, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("Error to retrieve in cluster config: %v", err)
	}
	kClient, err := kubernetes.NewForConfig(clusterConfig)
	if err != nil {
		return fmt.Errorf("Error to retrieve kubernetes client: %v", err)
	}
	vs.kClient = kClient
	return nil
}

// CreateVolumeFromSnapshot creates a new volume in the specified
// availability zone, initialized from the provided snapshot,
// and with the specified type and IOPS (if using provisioned IOPS).
func (vs *VolumeSnapshotter) CreateVolumeFromSnapshot(snapshotID, volumeType, volumeAZ string, iops *int64) (string, error) {
	vs.Log.Infof("CreateVolumeFromSnapshot for snapshotID: %s, volumeType: %s, volumeAZ: %s, iops: %v", snapshotID, volumeType, volumeAZ, *iops)
	// TODO
	return "", nil
}

// GetVolumeInfo returns the type and IOPS (if using provisioned IOPS) for
// the specified volume in the given availability zone.
func (vs *VolumeSnapshotter) GetVolumeInfo(volumeID, volumeAZ string) (string, *int64, error) {
	vs.Log.Infof("GetVolumeInfo for volumeID: %s, volumeAZ: %s", volumeID, volumeAZ)
	pv, err := vs.kClient.CoreV1().PersistentVolumes().Get(context.TODO(), volumeID, metav1.GetOptions{})
	if err != nil {
		return "", nil, fmt.Errorf("Unable to retrieve pv from %s", volumeID)
	}

	if pv.Spec.CSI == nil {
		return "", nil, fmt.Errorf("Unable to retrieve csi Spec from pv %+v", pv)
	}
	if pv.Spec.CSI.FSType == "" {
		return "", nil, fmt.Errorf("Unable to retrieve fs type from pv %+v", pv)
	}
	return pv.Spec.CSI.FSType, nil, nil
}

// CreateSnapshot creates a snapshot of the specified volume, and applies any provided
// set of tags to the snapshot.
func (vs *VolumeSnapshotter) CreateSnapshot(volumeID, volumeAZ string, tags map[string]string) (string, error) {
	vs.Log.Infof("CreateSnapshot for volumeID: %s, volumeAZ: %s, tags: %v", volumeID, volumeAZ, tags)

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/volumes/%s?action=snapshotCreate", backendServiceURL, volumeID),
		strings.NewReader(`{}`))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Unexpected response code: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	type Response struct {
		ID   string `json:"id,omitempty"`
		Name string `json:"name,omitempty"`
	}
	r := &Response{}
	if err := json.Unmarshal(body, r); err != nil {
		return "", err
	}

	snapshotID := r.ID
	if snapshotID == "" && r.Name != "" {
		snapshotID = r.Name
	}
	if snapshotID == "" {
		return "", fmt.Errorf("Empty snapshot ID")
	}
	vs.Log.Infof("CreateSnapshot for volumeID %s with snapshotID: %s", volumeID, snapshotID)
	return snapshotID, nil
}

// DeleteSnapshot deletes the specified volume snapshot.
func (vs *VolumeSnapshotter) DeleteSnapshot(snapshotID string) error {
	vs.Log.Infof("DeleteSnapshot for snapshotID: %s", snapshotID)

	// List all volumes to find the snapshotID belongs to which volume
	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s/v1/volumes", backendServiceURL),
		strings.NewReader(""))
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected response code: %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	type Actions struct {
		SnapshotGet  string `json:"snapshotGet,omitempty"`
		SnapshotList string `json:"snapshotList,omitempty"`
	}
	type Data struct {
		Actions Actions `json:"actions,omitempty"`
		ID      string  `json:"id,omitempty"`
		Name    string  `json:"name,omitempty"`
	}
	type VolumeResponse struct {
		Data []Data `json:"data,omitempty"`
	}
	volumeResponse := &VolumeResponse{}
	if err := json.Unmarshal(body, volumeResponse); err != nil {
		return err
	}

	volumeID := ""
	for _, data := range volumeResponse.Data {
		snapshotGetURL := data.Actions.SnapshotGet
		req, err := http.NewRequest(
			http.MethodGet,
			snapshotGetURL,
			strings.NewReader(fmt.Sprintf("{\"name\":\"%s\"}", snapshotID)))
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			continue
		}
		if data.ID != "" {
			volumeID = data.ID
		} else if data.Name != "" {
			volumeID = data.Name
		}
		break
	}

	if volumeID == "" {
		return fmt.Errorf("Cannot find the volume for snapshotID: %s", snapshotID)
	}

	// Find the volumeID for snapshotID, delete it
	req, err = http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/v1/volumes/%s?action=snapshotDelete", backendServiceURL, volumeID),
		strings.NewReader(fmt.Sprintf("{\"name\":\"%s\"}", snapshotID)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected response code: %d", resp.StatusCode)
	}
	vs.Log.Infof("DeleteSnapshot for snapshotID %s on volumeID %s", snapshotID, volumeID)
	return nil
}

// GetVolumeID returns the specific identifier for the PersistentVolume.
func (vs *VolumeSnapshotter) GetVolumeID(unstructuredPV runtime.Unstructured) (string, error) {
	pv := new(corev1.PersistentVolume)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPV.UnstructuredContent(), pv); err != nil {
		return "", errors.WithStack(err)
	}

	if pv.Spec.CSI == nil {
		return "", fmt.Errorf("Unable to retrieve csi spec from pv %+v", pv)
	}
	if pv.Spec.CSI.VolumeHandle == "" {
		return "", fmt.Errorf("Unable to retrieve volume handle from pv %+v", pv)
	}
	return pv.Spec.CSI.VolumeHandle, nil
}

// SetVolumeID sets the specific identifier for the PersistentVolume.
func (vs *VolumeSnapshotter) SetVolumeID(unstructuredPV runtime.Unstructured, volumeID string) (runtime.Unstructured, error) {
	pv := new(corev1.PersistentVolume)
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredPV.UnstructuredContent(), pv); err != nil {
		return nil, errors.WithStack(err)
	}

	if pv.Spec.CSI == nil {
		return nil, fmt.Errorf("Unable to retrieve csi spec from pv %+v", pv)
	}

	pv.Spec.CSI.VolumeHandle = volumeID
	res, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pv)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &unstructured.Unstructured{Object: res}, nil
}
