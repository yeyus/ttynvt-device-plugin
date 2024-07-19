package wrapper

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/google/uuid"

	"k8s.io/klog/v2"
)

const MAJOR = 199

type TTYVNTInstance struct {
	UUID       uuid.UUID
	DeviceName string
	Endpoint   string
	CancelFunc context.CancelFunc
	Command    *exec.Cmd
	StartTime  time.Time
	observers  map[InstanceObserver]struct{}
}

// Sample command: ttynvt -M 199 -d -m 6 -n TTYNVT0 -S 127.0.0.1:5099
func NewTTYVNTInstance(ctx context.Context, deviceName string, minor int, server string, port int) *TTYVNTInstance {
	endpoint := fmt.Sprintf("%s:%d", server, port)
	endpointArg := fmt.Sprintf("-S %s", endpoint)
	args := []string{"-d", fmt.Sprintf("-M %d", MAJOR), fmt.Sprintf("-m %d", minor), fmt.Sprintf("-n %s", deviceName), endpointArg}

	klog.Infof("[%s] with args %s", deviceName, args)
	instanceCtx, cancelFn := context.WithCancel(ctx)
	cmd := exec.CommandContext(instanceCtx, "ttynvt", args...)

	instance := &TTYVNTInstance{
		UUID:       uuid.New(),
		DeviceName: deviceName,
		Endpoint:   endpoint,
		CancelFunc: cancelFn,
		Command:    cmd,
		observers:  make(map[InstanceObserver]struct{}),
	}

	return instance
}

func (ti *TTYVNTInstance) Register(o InstanceObserver) {
	ti.observers[o] = struct{}{}
}

func (ti *TTYVNTInstance) Unregister(o InstanceObserver) {
	delete(ti.observers, o)
}

func (ti *TTYVNTInstance) Notify(e InstanceEvent) {
	for o := range ti.observers {
		o.OnNotify(e)
	}
}

func processWatch(ti *TTYVNTInstance) {
	err := ti.Command.Wait()
	if err != nil {
		klog.Errorf("[%s] Exited with error %s", ti.DeviceName, err)
	} else {
		klog.Infof("[%s] is done", ti.DeviceName)
	}
	ti.Notify(NewExitInstanceEvent(ti, err))
}

func (ti *TTYVNTInstance) Start() error {
	stdout, err := ti.Command.StdoutPipe()
	if err != nil {
		klog.Error(err)
		return err
	}

	stderr, err := ti.Command.StderrPipe()
	if err != nil {
		klog.Error(err)
		return err
	}

	stdoutScanner := bufio.NewScanner(stdout)
	go func() {
		for stdoutScanner.Scan() {
			klog.Infof("[stdout-%s] %s", ti.DeviceName, stdoutScanner.Text())
		}
		if err := stdoutScanner.Err(); err != nil {
			klog.Errorf("[error-%s] %s", ti.DeviceName, err)
		}
	}()

	stderrScanner := bufio.NewScanner(stderr)
	go func() {
		for stderrScanner.Scan() {
			klog.Errorf("[stderr-%s] %s", ti.DeviceName, stderrScanner.Text())
		}
		if err := stderrScanner.Err(); err != nil {
			klog.Errorf("[error-%s] %s", ti.DeviceName, err)
		}
	}()

	err = ti.Command.Start()
	ti.Notify(NewStartInstanceEvent(ti))
	ti.StartTime = time.Now()
	klog.Infof("[%s] Started virtual port with uuid=%s and PID %d", ti.DeviceName, ti.UUID.String(), ti.Command.Process.Pid)
	if err != nil {
		klog.Errorf("Error while instantiating ttynvt %s: %s", ti.DeviceName, err)
		return err
	}

	go processWatch(ti)

	return nil
}

func (ti *TTYVNTInstance) Kill() {
	if ti.CancelFunc == nil {
		return
	}

	klog.Infof("[%s] Killing virtual port with uuid=%s and PID %d", ti.DeviceName, ti.UUID.String(), ti.Command.Process.Pid)
	ti.CancelFunc()
}
