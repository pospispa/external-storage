package sharedfilesystems

import (
	"reflect"
	"testing"

	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/kubernetes/pkg/volume"

	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
)

func TestPrepareCreateRequest(t *testing.T) {
	functionUnderTest := "prepareCreateRequestv2"
	zonesForSCMultiZoneTestCase := "nova1, nova2, nova3"
	setOfZonesForSCMultiZoneTestCase, _ := zonesToSet(zonesForSCMultiZoneTestCase)
	pvcNameForSCMultiZoneTestCase := "pvc"
	expectedResultForSCMultiZoneTestCase := volume.ChooseZoneForVolume(setOfZonesForSCMultiZoneTestCase, pvcNameForSCMultiZoneTestCase)
	// First part: want no error
	succCases := []struct {
		volumeOptions controller.VolumeOptions
		storageSize   string
		want          shares.CreateOpts
	}{
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.Quantity{},
							},
						},
					},
				},
				Parameters: map[string]string{},
			},
			storageSize: "2G",
			want: shares.CreateOpts{
				ShareProto: ProtocolNFS,
				Size:       2,
			},
		},
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.Quantity{},
							},
						},
					},
				},
				Parameters: map[string]string{ZonesSCParamName: "nova"},
			},
			storageSize: "2G",
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				AvailabilityZone: "nova",
				Size:             2,
			},
		},
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.Quantity{},
							},
						},
					},
				},
				Parameters: map[string]string{"ZoNes": "nova"},
			},
			storageSize: "2G",
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				AvailabilityZone: "nova",
				Size:             2,
			},
		},
		// Will very probably start failing if the func volume.ChooseZoneForVolume is replaced by another function in the implementation
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.Quantity{},
							},
						},
					},
				},
				Parameters: map[string]string{ZonesSCParamName: zonesForSCMultiZoneTestCase},
			},
			storageSize: "2G",
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				AvailabilityZone: expectedResultForSCMultiZoneTestCase,
				Size:             2,
			},
		},
	}
	for _, succCase := range succCases {
		if quantity, err := resource.ParseQuantity(succCase.storageSize); err != nil {
			t.Errorf("Failed to parse storage size (%v): %v", succCase.storageSize, err)
			continue
		} else {
			succCase.volumeOptions.PVC.Spec.Resources.Requests[v1.ResourceStorage] = quantity
		}
		if request, err := prepareCreateRequest(succCase.volumeOptions); err != nil {
			t.Errorf("%v(%v) RETURNED (%v, %v), WANT (%v, %v)", functionUnderTest, succCase.volumeOptions, request, err, succCase.want, nil)
		} else if !reflect.DeepEqual(request, succCase.want) {
			t.Errorf("%v(%v) RETURNED (%v, %v), WANT (%v, %v)", functionUnderTest, succCase.volumeOptions, request, err, succCase.want, nil)
		}
	}

	// Second part: want an error
	errCases := []struct {
		volumeOptions controller.VolumeOptions
		storageSize   string
	}{
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.Quantity{},
							},
						},
					},
				},
				Parameters: map[string]string{"foo": "bar"},
			},
			storageSize: "2G",
		},
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.Quantity{},
							},
						},
					},
				},
				Parameters: map[string]string{},
			},
			storageSize: "2Gi",
		},
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.Quantity{},
							},
						},
					},
				},
				Parameters: map[string]string{},
			},
			storageSize: "0G",
		},
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.Quantity{},
							},
						},
					},
				},
				Parameters: map[string]string{},
			},
			storageSize: "-1G",
		},
	}
	for _, errCase := range errCases {
		if quantity, err := resource.ParseQuantity(errCase.storageSize); err != nil {
			t.Errorf("Failed to parse storage size (%v): %v", errCase.storageSize, err)
			continue
		} else {
			errCase.volumeOptions.PVC.Spec.Resources.Requests[v1.ResourceStorage] = quantity
		}
		if request, err := prepareCreateRequest(errCase.volumeOptions); err == nil {
			t.Errorf("%v(%v) RETURNED (%v, %v), WANT (%v, %v)", functionUnderTest, errCase.volumeOptions, request, err, "N/A", "an error")
		}
	}

	// Third part: want an error
	errCasesStorageSizeNotConfigured := []controller.VolumeOptions{
		{
			PersistentVolumeReclaimPolicy: "Delete",
			PVName: "pv",
			PVC: &v1.PersistentVolumeClaim{
				ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
				Spec:       v1.PersistentVolumeClaimSpec{},
			},
			Parameters: map[string]string{},
		},
		{
			PersistentVolumeReclaimPolicy: "Delete",
			PVName: "pv",
			PVC: &v1.PersistentVolumeClaim{
				ObjectMeta: v1.ObjectMeta{Name: "pvc", Namespace: "foo"},
				Spec: v1.PersistentVolumeClaimSpec{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU: resource.Quantity{},
						},
					},
				},
			},
			Parameters: map[string]string{},
		},
	}
	for _, errCase := range errCasesStorageSizeNotConfigured {
		if request, err := prepareCreateRequest(errCase); err == nil {
			t.Errorf("%v(%v) RETURNED (%v, %v), WANT (%v, %v)", functionUnderTest, errCase, request, err, "N/A", "an error")
		}
	}
}
