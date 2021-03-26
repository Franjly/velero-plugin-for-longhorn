### velero-plugin

Longhorn plugin for Velero. To take snapshots of Longhorn volumes through Velero you need to install and configure the Longhorn VolumeSnapshotter plugin.

#### Installation

1. Create a VolumeSnapshotLocation CR for Longhorn VolumeSnapshotter
   ```yaml
   apiVersion: velero.io/v1
   kind: VolumeSnapshotLocation
   metadata:
     name: longhorn
     namespace: velero
   spec:
     provider: longhorn.io/longhorn
   ```
   or
   ```shell
   velero snapshot-location create longhorn \
       --provider longhorn.io/longhorn
   ```

2. Add Longhorn plugin to Velero server
   ```shell
   velero plugin add quay.io/jenting/velero-plugin:main \
       --image-pull-policy Always
   ```

#### Create Backup

```shell
velero backup create <backup-name> \
    --volume-snapshot-locations longhorn
```

#### Delete Backup

```shell
velero backup delete <backup-name>
```
