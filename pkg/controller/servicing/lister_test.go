package servicing

import (
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sapcc/kubernikus/pkg/api/models"
	v1 "github.com/sapcc/kubernikus/pkg/apis/kubernikus/v1"
	"github.com/sapcc/kubernikus/pkg/controller/nodeobservatory"

	"k8s.io/apimachinery/pkg/runtime"
)

type MockNodeListerFactory struct {
	mock.Mock
}

func (m *MockNodeListerFactory) Make(k *v1.Kluster) LifeCycler {
	return m.Called(k).Get(0).(LifeCycler)
}

func NewFakeNodeLister(t *testing.T, logger log.Logger, kluster *v1.Kluster, nodes []runtime.Object) Lister {
	kl, _ := nodeobservatory.NewFakeController(kluster, nodes...).GetListerForKluster(kluster)

	var lister Lister
	lister = &NodeLister{
		Logger:        logger,
		Kluster:       kluster,
		Lister:        kl,
		CoreOSVersion: NewFakeLatestCoreOSVersion(t, "2023.4.0"),
	}

	lister = &LoggingLister{
		Lister: lister,
		Logger: logger,
	}

	return lister
}

func NewFakeKlusterForListerTests() (*v1.Kluster, []runtime.Object) {
	return NewFakeKluster(
		&FakeKlusterOptions{
			Phase:       models.KlusterPhaseRunning,
			LastService: nil,
			NodePools: []FakeNodePoolOptions{
				FakeNodePoolOptions{
					AllowReboot:         true,
					AllowReplace:        true,
					NodeOSOutdated:      true,
					NodeKubeletOutdated: true,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         false,
					AllowReplace:        true,
					NodeOSOutdated:      true,
					NodeKubeletOutdated: true,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         true,
					AllowReplace:        false,
					NodeOSOutdated:      true,
					NodeKubeletOutdated: true,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         true,
					AllowReplace:        true,
					NodeOSOutdated:      false,
					NodeKubeletOutdated: true,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         true,
					AllowReplace:        true,
					NodeOSOutdated:      true,
					NodeKubeletOutdated: false,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         false,
					AllowReplace:        false,
					NodeOSOutdated:      true,
					NodeKubeletOutdated: true,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         false,
					AllowReplace:        true,
					NodeOSOutdated:      false,
					NodeKubeletOutdated: true,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         false,
					AllowReplace:        true,
					NodeOSOutdated:      true,
					NodeKubeletOutdated: false,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         true,
					AllowReplace:        false,
					NodeOSOutdated:      false,
					NodeKubeletOutdated: true,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         true,
					AllowReplace:        false,
					NodeOSOutdated:      true,
					NodeKubeletOutdated: false,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         true,
					AllowReplace:        true,
					NodeOSOutdated:      false,
					NodeKubeletOutdated: false,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         true,
					AllowReplace:        false,
					NodeOSOutdated:      false,
					NodeKubeletOutdated: false,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         false,
					AllowReplace:        true,
					NodeOSOutdated:      false,
					NodeKubeletOutdated: false,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         false,
					AllowReplace:        false,
					NodeOSOutdated:      true,
					NodeKubeletOutdated: false,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         false,
					AllowReplace:        false,
					NodeOSOutdated:      false,
					NodeKubeletOutdated: true,
					Size:                1,
				},
				FakeNodePoolOptions{
					AllowReboot:         false,
					AllowReplace:        false,
					NodeOSOutdated:      false,
					NodeKubeletOutdated: false,
					NodeHealthy:         true,
					Size:                1,
				},
			},
		},
	)
}
func TestServicingListertAll(t *testing.T) {
	kluster, nodes := NewFakeKlusterForListerTests()
	lister := NewFakeNodeLister(t, TestLogger(), kluster, nodes)
	assert.Len(t, lister.All(), 16)
}

func TestServicingListerRequiringReboot(t *testing.T) {
	kluster, nodes := NewFakeKlusterForListerTests()
	lister := NewFakeNodeLister(t, TestLogger(), kluster, nodes)
	assert.Len(t, lister.RequiringReboot(), 4)
}

func TestServicingListerRequiringReplacement(t *testing.T) {
	kluster, nodes := NewFakeKlusterForListerTests()
	lister := NewFakeNodeLister(t, TestLogger(), kluster, nodes)
	assert.Len(t, lister.RequiringReplacement(), 4)
}

func TestServicingListerNotReady(t *testing.T) {
	kluster, nodes := NewFakeKlusterForListerTests()
	lister := NewFakeNodeLister(t, TestLogger(), kluster, nodes)
	assert.Len(t, lister.NotReady(), 15)
}
