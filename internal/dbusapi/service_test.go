package dbusapi

import (
	"reflect"
	"testing"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

// TestRequestWaitClose tests Wait and Close of Request
func TestRequestWaitClose(_ *testing.T) {
	// test closing
	r := Request{
		Name: "test1",
		wait: make(chan struct{}),
		done: make(chan struct{}),
	}
	go func() {
		r.Close()
	}()
	r.Wait()

	// test aborting
	done := make(chan struct{})
	r = Request{
		Name: "test2",
		wait: make(chan struct{}),
		done: done,
	}
	go func() {
		close(done)
	}()
	r.Wait()
}

// TestDaemonConnect tests Connect of daemon
func TestDaemonConnect(t *testing.T) {
	// create daemon
	requests := make(chan *Request)
	done := make(chan struct{})
	daemon := daemon{
		requests: requests,
		done:     done,
	}

	// run connect and get results
	cookie, host, connectURL, fingerprint, resolve :=
		"cookie", "host", "connectURL", "fingerprint", "resolve"
	want := &Request{
		Name:       RequestConnect,
		Parameters: []any{cookie, host, connectURL, fingerprint, resolve},
		done:       done,
	}
	got := &Request{}
	go func() {
		r := <-requests
		got = r
		r.Close()
	}()
	err := daemon.Connect("sender", cookie, host, connectURL, fingerprint, resolve)
	if err != nil {
		t.Error(err)
	}

	// check results
	if got.Name != want.Name ||
		!reflect.DeepEqual(got.Parameters, want.Parameters) ||
		!reflect.DeepEqual(got.Results, want.Results) ||
		got.Error != want.Error ||
		got.done != want.done {
		// not equal
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestDaemonDisconnect tests Disconnect of daemon
func TestDaemonDisconnect(t *testing.T) {
	// create daemon
	requests := make(chan *Request)
	done := make(chan struct{})
	daemon := daemon{
		requests: requests,
		done:     done,
	}

	// run disconnect and get results
	want := &Request{
		Name: RequestDisconnect,
		done: done,
	}
	got := &Request{}
	go func() {
		r := <-requests
		got = r
		r.Close()
	}()
	err := daemon.Disconnect("sender")
	if err != nil {
		t.Error(err)
	}

	// check results
	if got.Name != want.Name ||
		!reflect.DeepEqual(got.Parameters, want.Parameters) ||
		!reflect.DeepEqual(got.Results, want.Results) ||
		got.Error != want.Error ||
		got.done != want.done {
		// not equal
		t.Errorf("got %v, want %v", got, want)
	}
}

// testConn implements the dbusConn interface for testing
type testConn struct{}

func (tc *testConn) Close() error {
	return nil
}

func (tc *testConn) Export(any, dbus.ObjectPath, string) error {
	return nil
}

func (tc *testConn) RequestName(string, dbus.RequestNameFlags) (dbus.RequestNameReply, error) {
	return dbus.RequestNameReplyPrimaryOwner, nil
}

// testProperties implements the propProperties interface for testing
type testProperties struct {
	props map[string]any
}

func (tp *testProperties) Introspection(string) []introspect.Property {
	return nil
}

func (tp *testProperties) SetMust(_, property string, v any) {
	if tp.props == nil {
		// props not set, skip
		return
	}

	// ignore iface, map property to value
	tp.props[property] = v
}

// TestServiceStartStop tests Start and Stop of Service
func TestServiceStartStop(_ *testing.T) {
	dbusConnectSystemBus = func(opts ...dbus.ConnOption) (dbusConn, error) {
		return &testConn{}, nil
	}
	propExport = func(conn dbusConn, path dbus.ObjectPath, props prop.Map) (propProperties, error) {
		return &testProperties{}, nil
	}
	s := NewService()
	s.Start()
	s.Stop()
}

// TestServiceRequests tests Requests of Service
func TestServiceRequests(t *testing.T) {
	s := NewService()
	want := s.requests
	got := s.Requests()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestServiceSetProperty tests SetProperty of Service
func TestServiceSetProperty(t *testing.T) {
	dbusConnectSystemBus = func(opts ...dbus.ConnOption) (dbusConn, error) {
		return &testConn{}, nil
	}
	properties := &testProperties{props: make(map[string]any)}
	propExport = func(conn dbusConn, path dbus.ObjectPath, props prop.Map) (propProperties, error) {
		return properties, nil
	}
	s := NewService()
	s.Start()

	propName := "test-property"
	want := "test-value"

	s.SetProperty(propName, want)
	s.Stop()

	got := properties.props[propName]
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

// TestNewService tests NewService
func TestNewService(t *testing.T) {
	s := NewService()
	empty := &Service{}
	if reflect.DeepEqual(s, empty) {
		t.Errorf("got empty, want not empty")
	}
}
