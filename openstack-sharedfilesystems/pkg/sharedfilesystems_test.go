package sharedfilesystems

import (
	"reflect"
	"testing"

	"k8s.io/client-go/pkg/api/resource"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/types"
	"k8s.io/kubernetes/pkg/util/sets"
	"k8s.io/kubernetes/pkg/volume"

	"github.com/gophercloud/gophercloud/openstack/sharedfilesystems/v2/shares"
	"github.com/kubernetes-incubator/external-storage/lib/controller"
)

func mockGetAllZones() (sets.String, error) {
	ret := sets.String{"nova1": sets.Empty{}, "nova2": sets.Empty{}, "nova3": sets.Empty{}}
	return ret, nil
}

func TestPrepareCreateRequest(t *testing.T) {
	functionUnderTest := "PrepareCreateRequestv2"

	fakeUID := types.UID("unique-uid")
	fakeShareName := "pvc-" + string(fakeUID)
	fakePVCName := "pvc"
	fakeNamespace := "foo"

	zonesForSCMultiZoneTestCase := "nova1, nova2, nova3"
	setOfZonesForSCMultiZoneTestCase, _ := zonesToSet(zonesForSCMultiZoneTestCase)
	pvcNameForSCMultiZoneTestCase := "pvc"
	expectedResultForSCMultiZoneTestCase := volume.ChooseZoneForVolume(setOfZonesForSCMultiZoneTestCase, pvcNameForSCMultiZoneTestCase)
	pvcNameForSCNoZonesSpecifiedTestCase := "pvc"
	allClusterZonesForSCNoZonesSpecifiedTestCase, _ := mockGetAllZones()
	expectedResultForSCNoZonesSpecifiedTestCase := volume.ChooseZoneForVolume(allClusterZonesForSCNoZonesSpecifiedTestCase, pvcNameForSCNoZonesSpecifiedTestCase)
	// First part: want no error
	succCases := []struct {
		volumeOptions controller.VolumeOptions
		storageSize   string
		want          shares.CreateOpts
	}{
		// Will very probably start failing if the func volume.ChooseZoneForVolume is replaced by another function in the implementation
		{
			volumeOptions: controller.VolumeOptions{
				PersistentVolumeReclaimPolicy: "Delete",
				PVName: "pv",
				PVC: &v1.PersistentVolumeClaim{
					ObjectMeta: v1.ObjectMeta{
						Name:      pvcNameForSCNoZonesSpecifiedTestCase,
						Namespace: fakeNamespace,
						UID:       fakeUID},
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
					ObjectMeta: v1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
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
					ObjectMeta: v1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
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
					ObjectMeta: v1.ObjectMeta{
						Name:      pvcNameForSCMultiZoneTestCase,
						Namespace: fakeNamespace,
						UID:       fakeUID},
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
					ObjectMeta: v1.ObjectMeta{
						Name:      fakePVCName,
						Namespace: fakeNamespace,
						UID:       fakeUID},
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadOnlyMany},
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
				Name:             fakeShareName,
				AvailabilityZone: "nova",
				Size:             2,
			},
		},
	}
	for i, succCase := range succCases {
		if quantity, err := resource.ParseQuantity(succCase.storageSize); err != nil {
			t.Errorf("Test case %v: Failed to parse storage size (%v): %v", i, succCase.storageSize, err)
			continue
		} else {
			succCase.volumeOptions.PVC.Spec.Resources.Requests[v1.ResourceStorage] = quantity
		}
		tags := make(map[string]string)
		tags[CloudVolumeCreatedForClaimNamespaceTag] = fakeNamespace
		tags[CloudVolumeCreatedForClaimNameTag] = succCase.volumeOptions.PVC.Name
		tags[CloudVolumeCreatedForVolumeNameTag] = succCase.want.Name
		succCase.want.Metadata = tags
		if request, err := PrepareCreateRequest(succCase.volumeOptions, mockGetAllZones); err != nil {
			t.Errorf("Test case %v: %v(%v) RETURNED (%v, %v), WANT (%v, %v)", i, functionUnderTest, succCase.volumeOptions, request, err, succCase.want, nil)
		} else if !reflect.DeepEqual(request, succCase.want) {
			t.Errorf("Test case %v: %v(%v) RETURNED (%v, %v), WANT (%v, %v)", i, functionUnderTest, succCase.volumeOptions, request, err, succCase.want, nil)
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
		if request, err := PrepareCreateRequest(errCase.volumeOptions, mockGetAllZones); err == nil {
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
		if request, err := PrepareCreateRequest(errCase, mockGetAllZones); err == nil {
			t.Errorf("%v(%v) RETURNED (%v, %v), WANT (%v, %v)", functionUnderTest, errCase, request, err, "N/A", "an error")
		}
	}
}

const (
	validPath              = "ip://directory"
	preferredPath          = "ip://preferred/directory"
	emptyPath              = ""
	spacesOnlyPath         = "  	  "
	shareExportLocationID1 = "123456-1"
	shareExportLocationID2 = "1234567-1"
	shareExportLocationID3 = "1234567-2"
	shareExportLocationID4 = "7654321-1"
	shareID1               = "123456"
	shareID2               = "1234567"
)

func TestChooseExportLocationSuccess(t *testing.T) {
	tests := []struct {
		testCaseName string
		locs         []shares.ExportLocation
		want         shares.ExportLocation
	}{
		{
			testCaseName: "Match first item:",
			locs: []shares.ExportLocation{
				{
					Path:            validPath,
					ShareInstanceID: shareID1,
					IsAdminOnly:     false,
					ID:              shareExportLocationID1,
					Preferred:       false,
				},
			},
			want: shares.ExportLocation{
				Path:            validPath,
				ShareInstanceID: shareID1,
				IsAdminOnly:     false,
				ID:              shareExportLocationID1,
				Preferred:       false,
			},
		},
		{
			testCaseName: "Match preferred location:",
			locs: []shares.ExportLocation{
				{
					Path:            validPath,
					ShareInstanceID: shareID2,
					IsAdminOnly:     false,
					ID:              shareExportLocationID2,
					Preferred:       false,
				},
				{
					Path:            preferredPath,
					ShareInstanceID: shareID2,
					IsAdminOnly:     false,
					ID:              shareExportLocationID3,
					Preferred:       true,
				},
			},
			want: shares.ExportLocation{
				Path:            preferredPath,
				ShareInstanceID: shareID2,
				IsAdminOnly:     false,
				ID:              shareExportLocationID3,
				Preferred:       true,
			},
		},
		{
			testCaseName: "Match first not-preferred location that matches shareID:",
			locs: []shares.ExportLocation{
				{
					Path:            validPath,
					ShareInstanceID: shareID2,
					IsAdminOnly:     false,
					ID:              shareExportLocationID2,
					Preferred:       false,
				},
				{
					Path:            preferredPath,
					ShareInstanceID: shareID2,
					IsAdminOnly:     false,
					ID:              shareExportLocationID3,
					Preferred:       false,
				},
			},
			want: shares.ExportLocation{
				Path:            validPath,
				ShareInstanceID: shareID2,
				IsAdminOnly:     false,
				ID:              shareExportLocationID2,
				Preferred:       false,
			},
		},
	}

	for _, tt := range tests {
		if got, err := ChooseExportLocation(tt.locs); err != nil {
			t.Errorf("%q ChooseExportLocation(%v) = (%v, %q) want (%v, nil)", tt.testCaseName, tt.locs, got, err.Error(), tt.want)
		} else if !reflect.DeepEqual(tt.want, got) {
			t.Errorf("%q ChooseExportLocation(%v) = (%v, nil) want (%v, nil)", tt.testCaseName, tt.locs, got, tt.want)
		}
	}
}

func TestChooseExportLocationNotFound(t *testing.T) {
	tests := []struct {
		testCaseName string
		locs         []shares.ExportLocation
	}{
		{
			testCaseName: "Empty slice:",
			locs:         []shares.ExportLocation{},
		},
		{
			testCaseName: "Locations for admins only:",
			locs: []shares.ExportLocation{
				{
					Path:            validPath,
					ShareInstanceID: shareID1,
					IsAdminOnly:     true,
					ID:              shareExportLocationID1,
					Preferred:       false,
				},
			},
		},
		{
			testCaseName: "Preferred locations for admins only:",
			locs: []shares.ExportLocation{
				{
					Path:            validPath,
					ShareInstanceID: shareID1,
					IsAdminOnly:     true,
					ID:              shareExportLocationID1,
					Preferred:       true,
				},
			},
		},
		{
			testCaseName: "Empty path:",
			locs: []shares.ExportLocation{
				{
					Path:            emptyPath,
					ShareInstanceID: shareID1,
					IsAdminOnly:     false,
					ID:              shareExportLocationID1,
					Preferred:       false,
				},
			},
		},
		{
			testCaseName: "Empty path in preferred location:",
			locs: []shares.ExportLocation{
				{
					Path:            emptyPath,
					ShareInstanceID: shareID1,
					IsAdminOnly:     false,
					ID:              shareExportLocationID1,
					Preferred:       true,
				},
			},
		},
		{
			testCaseName: "Path containing spaces only:",
			locs: []shares.ExportLocation{
				{
					Path:            spacesOnlyPath,
					ShareInstanceID: shareID1,
					IsAdminOnly:     false,
					ID:              shareExportLocationID1,
					Preferred:       false,
				},
			},
		},
		{
			testCaseName: "Preferred path containing spaces only:",
			locs: []shares.ExportLocation{
				{
					Path:            spacesOnlyPath,
					ShareInstanceID: shareID1,
					IsAdminOnly:     false,
					ID:              shareExportLocationID1,
					Preferred:       true,
				},
			},
		},
	}
	for _, tt := range tests {
		if got, err := ChooseExportLocation(tt.locs); err == nil {
			t.Errorf("%q ChooseExportLocation(%v) = (%v, nil) want (\"N/A\", \"an error\")", tt.testCaseName, tt.locs, got)
		}
	}
}
