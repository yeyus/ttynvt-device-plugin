package wrapper

import (
	"context"
	"errors"
	"fmt"

	"github.com/mergestat/timediff"
	"k8s.io/klog/v2"
)

const DEFAULT_RFC2217_PORT = 5099
const MINOR_SHIFT = 5

type TTYNVTManager struct {
	Instances []*TTYVNTInstance

	maxTTY int
}

func NewTTYNVTManager(maxTTY int) *TTYNVTManager {
	return &TTYNVTManager{
		Instances: make([]*TTYVNTInstance, maxTTY),

		maxTTY: maxTTY,
	}
}

func (m *TTYNVTManager) Create(service string) error {
	slot := m.nextAvailableSlot()
	if slot < 0 {
		klog.Errorf("failure while allocating slot for a new instances, max instances = %d", m.maxTTY)
		return errors.New("tty allocations exhausted")
	}

	deviceName := fmt.Sprintf("ttyNVT%d", slot)

	instance := NewTTYVNTInstance(context.TODO(), deviceName, MINOR_SHIFT+slot, service, DEFAULT_RFC2217_PORT)
	instance.Register(m)
	m.Instances[slot] = instance

	return instance.Start()
}

func (m *TTYNVTManager) OnNotify(e InstanceEvent) {
	klog.Infof("Notified: %v", e)
	switch e.Type {
	case EventExit:
		m.freeSlot(e.Origin)
	}
}

func (m *TTYNVTManager) Print() {
	klog.Infoln("\t SLOT\t UUID\t DEVICE\t ENDPOINT\t DURATION\t")
	for i, slot := range m.Instances {
		if slot == nil {
			klog.Infof("\t %d\t Empty", i)
		} else {
			relatime := timediff.TimeDiff(slot.StartTime)
			klog.Infof("\t %d\t %s\t %s\t %s\t %s\t", i, slot.UUID.String(), slot.DeviceName, slot.Endpoint, relatime)
		}
	}
}

// Returns the next available slot or -1 if no more slots are available
func (m *TTYNVTManager) nextAvailableSlot() int {
	for i := 0; i < m.maxTTY; i++ {
		if m.Instances[i] == nil {
			return i
		}
	}

	return -1
}

func (m *TTYNVTManager) freeSlot(instance *TTYVNTInstance) {
	for slot, inst := range m.Instances {
		if inst == nil {
			continue
		}

		if inst == instance {
			klog.Infof("[%s] removing instance %s from slot %d", inst.DeviceName, inst.UUID.String(), slot)
			m.Instances[slot] = nil
			return
		}
	}

	klog.Errorf("[Manager] Tried removing instance %s for port %s, but couldn't find it!", instance.UUID.String(), instance.DeviceName)
}
