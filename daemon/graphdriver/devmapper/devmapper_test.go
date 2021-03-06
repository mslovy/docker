// +build linux

package devmapper

import (
	"fmt"
	"testing"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/daemon/graphdriver/graphtest"
)

func init() {
	// Reduce the size the the base fs and loopback for the tests
	defaultDataLoopbackSize = 300 * 1024 * 1024
	defaultMetaDataLoopbackSize = 200 * 1024 * 1024
	defaultBaseFsSize = 300 * 1024 * 1024
	defaultUdevSyncOverride = true
	if err := graphtest.InitLoopbacks(); err != nil {
		panic(err)
	}
}

// This avoids creating a new driver for each test if all tests are run
// Make sure to put new tests between TestDevmapperSetup and TestDevmapperTeardown
func TestDevmapperSetup(t *testing.T) {
	graphtest.GetDriver(t, "devicemapper")
}

func TestDevmapperCreateEmpty(t *testing.T) {
	graphtest.DriverTestCreateEmpty(t, "devicemapper")
}

func TestDevmapperCreateBase(t *testing.T) {
	graphtest.DriverTestCreateBase(t, "devicemapper")
}

func TestDevmapperCreateSnap(t *testing.T) {
	graphtest.DriverTestCreateSnap(t, "devicemapper")
}

func TestDevmapperTeardown(t *testing.T) {
	graphtest.PutDriver(t)
}

func TestDevmapperReduceLoopBackSize(t *testing.T) {
	tenMB := int64(10 * 1024 * 1024)
	testChangeLoopBackSize(t, -tenMB, defaultDataLoopbackSize, defaultMetaDataLoopbackSize)
}

func TestDevmapperIncreaseLoopBackSize(t *testing.T) {
	tenMB := int64(10 * 1024 * 1024)
	testChangeLoopBackSize(t, tenMB, defaultDataLoopbackSize+tenMB, defaultMetaDataLoopbackSize+tenMB)
}

func testChangeLoopBackSize(t *testing.T, delta, expectDataSize, expectMetaDataSize int64) {
	driver := graphtest.GetDriver(t, "devicemapper").(*graphtest.Driver).Driver.(*graphdriver.NaiveDiffDriver).ProtoDriver.(*Driver)
	defer graphtest.PutDriver(t)
	// make sure data or metadata loopback size are the default size
	if s := driver.DeviceSet.Status(); s.Data.Total != uint64(defaultDataLoopbackSize) || s.Metadata.Total != uint64(defaultMetaDataLoopbackSize) {
		t.Fatalf("data or metadata loop back size is incorrect")
	}
	if err := driver.Cleanup(); err != nil {
		t.Fatal(err)
	}
	//Reload
	d, err := Init(driver.home, []string{
		fmt.Sprintf("dm.loopdatasize=%d", defaultDataLoopbackSize+delta),
		fmt.Sprintf("dm.loopmetadatasize=%d", defaultMetaDataLoopbackSize+delta),
	}, nil, nil)
	if err != nil {
		t.Fatalf("error creating devicemapper driver: %v", err)
	}
	driver = d.(*graphdriver.NaiveDiffDriver).ProtoDriver.(*Driver)
	if s := driver.DeviceSet.Status(); s.Data.Total != uint64(expectDataSize) || s.Metadata.Total != uint64(expectMetaDataSize) {
		t.Fatalf("data or metadata loop back size is incorrect")
	}
	if err := driver.Cleanup(); err != nil {
		t.Fatal(err)
	}
}
