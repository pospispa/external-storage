/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sharedfilesystems

import (
	"reflect"
	"testing"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/controller/volume/persistentvolume"
	"k8s.io/kubernetes/pkg/volume"
	"k8s.io/kubernetes/pkg/volume/util"

	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
)

const (
	fakeReclaimPolicy = "Delete"
	fakeZoneName      = "nova"
	fakePVName        = "pv"
)

func mockGetAllZones() (sets.String, error) {
	ret := sets.String{"nova1": sets.Empty{}, "nova2": sets.Empty{}, "nova3": sets.Empty{}}
	return ret, nil
}

func TestPrepareCreateRequest(t *testing.T) {
	functionUnderTest := "PrepareCreateRequest"

	fakeUID := types.UID("unique-uid")
	fakeShareName := "pvc-" + string(fakeUID)
	fakePVCName := "pvc"
	fakeNamespace := "foo"

	zonesForSCMultiZoneTestCase := "nova1, nova2, nova3"
	setOfZonesForSCMultiZoneTestCase, _ := util.ZonesToSet(zonesForSCMultiZoneTestCase)
	pvcNameForSCMultiZoneTestCase := "pvc"
	expectedResultForSCMultiZoneTestCase := volume.ChooseZoneForVolume(setOfZonesForSCMultiZoneTestCase, pvcNameForSCMultiZoneTestCase)
	pvcNameForSCNoZonesSpecifiedTestCase := "pvc"
	allClusterZonesForSCNoZonesSpecifiedTestCase, _ := mockGetAllZones()
	expectedResultForSCNoZonesSpecifiedTestCase := volume.ChooseZoneForVolume(allClusterZonesForSCNoZonesSpecifiedTestCase, pvcNameForSCNoZonesSpecifiedTestCase)
	succCaseStorageSize, _ := resource.ParseQuantity("2G")
	fakeShareTypeName := "default"
	// First part: want no error
	succCases := []struct {
		volumeOptions controller.VolumeOptions
		want          shares.CreateOpts
	}{
		// Will very probably start failing if the func volume.ChooseZoneForVolume is replaced by another function in the implementation
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvcNameForSCNoZonesSpecifiedTestCase,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: succCaseStorageSize,
							},
						},
					},
				},
				Parameters: map[string]string{},
			},
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				Name:             fakeShareName,
				AvailabilityZone: expectedResultForSCNoZonesSpecifiedTestCase,
				Size:             2,
			},
		},
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: succCaseStorageSize,
							},
						},
					},
				},
				Parameters: map[string]string{ZonesSCParamName: "nova"},
			},
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				Name:             fakeShareName,
				AvailabilityZone: "nova",
				Size:             2,
			},
		},
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: succCaseStorageSize,
							},
						},
					},
				},
				Parameters: map[string]string{"ZoNes": "nova"},
			},
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				Name:             fakeShareName,
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
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvcNameForSCMultiZoneTestCase,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: succCaseStorageSize,
							},
						},
					},
				},
				Parameters: map[string]string{ZonesSCParamName: zonesForSCMultiZoneTestCase},
			},
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				Name:             fakeShareName,
				AvailabilityZone: expectedResultForSCMultiZoneTestCase,
				Size:             2,
			},
		},
		// PVC accessModes parameters are being ignored.
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadOnlyMany},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: succCaseStorageSize,
							},
						},
					},
				},
				Parameters: map[string]string{ZonesSCParamName: "nova"},
			},
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				Name:             fakeShareName,
				AvailabilityZone: "nova",
				Size:             2,
			},
		},
		// In case the requested storage size in GB is not a whole number because of the chosen units the storage size in GB is rounded up to the nearest integer
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("2Gi"),
							},
						},
					},
				},
				Parameters: map[string]string{ZonesSCParamName: "nova"},
			},
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				Name:             fakeShareName,
				AvailabilityZone: "nova",
				Size:             3,
			},
		},
		// In case the requested storage size is not a whole number the storage size is rounded up to the nearest integer
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("2.2G"),
							},
						},
					},
				},
				Parameters: map[string]string{ZonesSCParamName: "nova"},
			},
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				Name:             fakeShareName,
				AvailabilityZone: "nova",
				Size:             3,
			},
		},
		// Optional parameter "type" is present in the Storage Class
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: fakeReclaimPolicy,
				PVName: fakePVName,
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("2.2G"),
							},
						},
					},
				},
				Parameters: map[string]string{ZonesSCParamName: fakeZoneName, TypeSCParamName: fakeShareTypeName},
			},
			want: shares.CreateOpts{
				ShareProto:       ProtocolNFS,
				Name:             fakeShareName,
				ShareType:        fakeShareTypeName,
				AvailabilityZone: fakeZoneName,
				Size:             3,
			},
		},
	}
	for i, succCase := range succCases {
		tags := make(map[string]string)
		tags[persistentvolume.CloudVolumeCreatedForClaimNamespaceTag] = fakeNamespace
		tags[persistentvolume.CloudVolumeCreatedForClaimNameTag] = succCase.volumeOptions.PVC.Name
		tags[persistentvolume.CloudVolumeCreatedForVolumeNameTag] = succCase.want.Name
		succCase.want.Metadata = tags
		if request, err := PrepareCreateRequest(succCase.volumeOptions, mockGetAllZones); err != nil {
			t.Errorf("Test case %v: %v(%v) RETURNED (%v, %v), WANT (%v, %v)", i, functionUnderTest, succCase.volumeOptions, request, err, succCase.want, nil)
		} else if !reflect.DeepEqual(request, succCase.want) {
			t.Errorf("Test case %v: %v(%v) RETURNED (%v, %v), WANT (%v, %v)", i, functionUnderTest, succCase.volumeOptions, request, err, succCase.want, nil)
		}
	}

	// Second part: want an error
	errCases := []struct {
		testCaseName  string
		volumeOptions controller.VolumeOptions
	}{
		{
			testCaseName: "unknown Storage Class option",
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("2G"),
							},
						},
					},
				},
				Parameters: map[string]string{"foo": "bar"},
			},
		},
		{
			testCaseName: "zero storage capacity",
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("0G"),
							},
						},
					},
				},
				Parameters: map[string]string{},
			},
		},
		{
			testCaseName: "negative storage capacity",
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{Name: "pvc", Namespace: "foo"},
					Spec: v1.PersistentVolumeClaimSpec{
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: resource.MustParse("-1G"),
							},
						},
					},
				},
				Parameters: map[string]string{},
			},
		},
	}
	for _, errCase := range errCases {
		if request, err := PrepareCreateRequest(errCase.volumeOptions, mockGetAllZones); err == nil {
			t.Errorf("Test case %q: %v(%v) RETURNED (%v, %v), WANT (%v, %v)", errCase.testCaseName, functionUnderTest, errCase.volumeOptions, request, err, "N/A", "an error")
		}
	}

	// Third part: want an error
	errCasesStorageSizeNotConfigured := []controller.VolumeOptions{
		{
			PersistentVolumeReclaimPolicy: "Delete",
			PVName: "pv",
			PVC: &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "pvc", Namespace: "foo"},
				Spec:       v1.PersistentVolumeClaimSpec{},
			},
			Parameters: map[string]string{},
		},
		{
			PersistentVolumeReclaimPolicy: "Delete",
			PVName: "pv",
			PVC: &v1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{Name: "pvc", Namespace: "foo"},
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
		if request, err := PrepareCreateRequest(errCase, mockGetAllZones); err == nil {
			t.Errorf("%v(%v) RETURNED (%v, %v), WANT (%v, %v)", functionUnderTest, errCase, request, err, "N/A", "an error")
		}
	}
}
