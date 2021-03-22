package main

import (
	"github.com/sirupsen/logrus"
	veleroplugin "github.com/vmware-tanzu/velero/pkg/plugin/framework"

	"github.com/ecatlabs/velero-plugin/pkg/plugin"
)

func main() {
	veleroplugin.NewServer().
		RegisterVolumeSnapshotter("longhorn.io/volume-snapshotter-plugin", newVolumeSnapshotterPlugin).
		Serve()
}

func newVolumeSnapshotterPlugin(logger logrus.FieldLogger) (interface{}, error) {
	return plugin.NewVolumeSnapshotter(logger), nil
}
