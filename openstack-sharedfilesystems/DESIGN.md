# OpenStack Manila External Provisioner


## Deployment Strategy into OpenShift
Consensus must be obtained on how the external provisioner will be deployed into OpenShift, however, no discussion has started yet.


## Security
Only supplemental group will be used.

### Supplemental Group
The [gidallocator package](https://github.com/wongma7/efs-provisioner/blob/master/pkg/gidallocator/allocator.go) will be used to allocate a GID for each provisioned share. The GID is given as a supplemental group to the process(es) running in a pod that mounted the provisioned share.
In addition, Manila access control for the provisioned share will be set to `ip 0.0.0.0` immediately after creation so that according to the Manila documentation the share can be mounted from any machine.

### Supplemental Group and Manila Access Control Feature
Supplemental group is used as described in the Supplemental group section.
Additionally, at the time of provisioning a certificate will be generated and access control to the provisioned share will be allowed only using the certificate.
A new (Flex Volume or out-of-tree or in-tree) plugin for mounting Manila share will have to be developed because the certificate must be used to mount the share.
This approach won't be implemented but might be considered as a future improvement.


## `gophercloud` Library
The [`gophercloud` library](https://github.com/gophercloud/gophercloud) will be used for communication with Manila API.

### Authentication to Manila Service
`gophercloud`library reads values for authentication from environment variables. The below environment variable combinations were successfully used for authentication against Keystone version Newton `2:10.0.0-0ubuntu1`:

```
OS_USERNAME=demo
OS_PASSWORD=openstack
OS_AUTH_URL=http://localhost:35357/v3
OS_DOMAIN_NAME=Default
OS_TENANT_NAME=demo
```

```
OS_USERID=7e22ce01934c47dcae0f90e96cdfcf03
OS_PASSWORD=openstack
OS_AUTH_URL=http://localhost:35357/v3
OS_TENANT_ID=ecbc0da9369f41e3a8a17e49a425ff2d
```

```
OS_USERNAME=demo
OS_PASSWORD=openstack
OS_AUTH_URL=http://localhost:35357/v3
OS_DOMAIN_ID=default
OS_TENANT_NAME=demo
```

```
OS_USERNAME=demo
OS_PASSWORD=openstack
OS_AUTH_URL=http://localhost:35357/v3
OS_DOMAIN_ID=default
OS_TENANT_ID=ecbc0da9369f41e3a8a17e49a425ff2d
```

Note: [an issue](https://github.com/gophercloud/gophercloud/issues/297) is logged because the `OS_PROJECT_NAME` and `OS_PROJECT_ID` environment variables are not taken into account.

#### Authentication Token Limited Validity
Authentication token created during authentication and used in follow-up API calls has limited validity and may expire after 1 hour. That means that the provisioner should **stop** working after 1 hour. Currently, there is no solution to this potential problem.

### Share Creation and Deletion
[`Create`, `Delete` and `Get` methods](https://github.com/gophercloud/gophercloud/blob/master/openstack/sharedfilesystems/v2/shares/requests.go) are already available and will be needed to create a new share or delete an existing share.

Important: the deletion shall be implemented in such a way that even though the share still contains some data the share is deleted. This behaviour may depend on the used Manila back-end.

### Access Control
Access control must be set to every newly created share, otherwise the share can't be mounted.
There're issues logged for [Manila API](https://github.com/gophercloud/gophercloud/issues/114) and [Manila shares](https://github.com/gophercloud/gophercloud/issues/129), however, nothing is implemented for the access control feature and there are no pull-requests for it.

Therefore, the below API calls have to be implemented:
- [Share export locations](https://developer.openstack.org/api-ref/shared-file-systems/#share-export-locations-since-api-v2-9) and maybe also [Share instance export locations](https://developer.openstack.org/api-ref/shared-file-systems/#share-instance-export-locations-since-api-v2-9).
- [Grant access](https://developer.openstack.org/api-ref/shared-file-systems/#grant-access) to a share.
- [List access rules](https://developer.openstack.org/api-ref/shared-file-systems/#list-access-rules) may be needed.


## Testing
[Testing pyramid](https://testing.googleblog.com/2015/04/just-say-no-to-more-end-to-end-tests.html) will be followed.

### Unit tests
Extensive unit tests that will be result of test driven development.

### Integration tests
No integration tests are planned.

### Kubernetes E2E tests
Assumption: OpenStack environment will be available.

Limitations to reduce the testing matrix:
- Only latest release of Kubernetes and Kubernetes master will be used for testing.
- In case the OpenStack version is Libery or above, i.e. contains Manila service, only this version of Manila service will be used for testing.
- In case the OpenStack version is Kilo or below, i.e. does not contain Manila service, the Newton version of Keystone, Rabbit MQ and Manila will be deployed separately into the OpenStack and used for testing.

To sum it up there will be automated E2E tests for the below combinations:
- Periodically: Kubernetes master with external provisioner master and with fixed version of Manila service.
- Periodically: Latest release of Kubernetes with external provisioner master and with fixed version of Manila service.
- Per request: Kubernetes master with external provisioner PR and with fixed version of Manila service.
- Per request: Latest release of Kubernetes with external provisioner master and with fixed version of Manila service.

### OpenShift E2E tests
TBD


## A Share Creation
Share creation consists of:
- [Create request](http://developer.openstack.org/api-ref/shared-file-systems/?expanded=create-share-detail#create-share) that either fails or results in a share being in state `creating`.
- `created` state waiting loop: because a successful share create request results in a `creating` share it is necessary to wait for a share to be created afterwards. So a waiting loop that periodically [shows the share status](http://developer.openstack.org/api-ref/shared-file-systems/?expanded=create-share-detail#show-share-details) after 1, 2, 4, 8, etc. seconds and waits until the status changes to `created` or the waiting timeouts (configurable timeout; default 180 seconds).
- Access control settings: depending on Product owner decision.


## Storage Class Example(s)
```
apiVersion: storage.k8s.io/v1beta1
kind: StorageClass
metadata:
  name: manilaNFSshare
provisioner: kubernetes.io/manila
parameters:
  zones: nova1, nova2, nova3
```
Optional parameter(s):
- `zones` a set of zones; one of the zones will be used as the `availability_zone` in the [Create request](http://developer.openstack.org/api-ref/shared-file-systems/?expanded=create-share-detail#create-share). In case the `zones` parameter is not specified the `availability_zone` in the [Create request](http://developer.openstack.org/api-ref/shared-file-systems/?expanded=create-share-detail#create-share) is filled with one of the zones that exist in the cluster.

Unavailable parameter(s):
- `share_proto` that is a mandatory parameter in the [Create request](http://developer.openstack.org/api-ref/shared-file-systems/?expanded=create-share-detail#create-share). The value of `NFS` will be always used.

[Create request](http://developer.openstack.org/api-ref/shared-file-systems/?expanded=create-share-detail#create-share) optional parameters that won't be supported in Storage Class:
- `share_type`
- `volume_type`


## PVC Example(s)
```
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: tinyshare
  annotations:
    "volume.beta.kubernetes.io/storage-class": "manilaNFSshare"
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 2G
```
Mandatory parameter(s):
- `storage` and the requested storage size must be whole integer number in GBs.

Ignored parameter(s):
- `accessModes` are ignored. A PV created on a PVC demand will contain all access modes supported by the corresponding filesystem specified in the corresponding Storage Class (note: currently, only NFS is supported that's why all ReadWriteOnce, ReadOnlyMany and ReadWriteMany access modes are filled into the PV).

[Create request](http://developer.openstack.org/api-ref/shared-file-systems/?expanded=create-share-detail#create-share) optional parameters that won't be supported in PVC:
- `name`
- `description`
- `display_name`
- `display_description`
- `snapshot_id`
- `is_public`
- `metadata`
- `share_network_id`
- `consistency_group_id`
