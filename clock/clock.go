package clock

/*
	#cgo LDFLAGS: -L../third_party/nim-c-library-guide/build/ -lclock
	#cgo LDFLAGS: -L../third_party/nim-c-library-guide -Wl,-rpath,../third_party/nim-c-library-guide/build/

	#include "../third_party/nim-c-library-guide/library/libclock.h"
	#include <stdio.h>
	#include <stdlib.h>

	extern void globalEventCallback(int ret, char* msg, size_t len, void* userData);

	typedef struct {
		int ret;
		char* msg;
		size_t len;
		void* ffiWg;
	} Resp;

	static void* allocResp(void* wg) {
		Resp* r = calloc(1, sizeof(Resp));
		r->ffiWg = wg;
		return r;
	}

	static void freeResp(void* resp) {
		if (resp != NULL) {
			free(resp);
		}
	}

	static char* getMyCharPtr(void* resp) {
		if (resp == NULL) {
			return NULL;
		}
		Resp* m = (Resp*) resp;
		return m->msg;
	}

	static size_t getMyCharLen(void* resp) {
		if (resp == NULL) {
			return 0;
		}
		Resp* m = (Resp*) resp;
		return m->len;
	}

	static int getRet(void* resp) {
		if (resp == NULL) {
			return 0;
		}
		Resp* m = (Resp*) resp;
		return m->ret;
	}

	// resp must be set != NULL in case interest on retrieving data from the callback
	void GoCallback(int ret, char* msg, size_t len, void* resp);

	static void* cGoNewClock(void* resp) {
		// We pass NULL because we are not interested in retrieving data from this callback
		void* ret = clock_new((ClockCallBack) GoCallback, resp);
		return ret;
	}

	static void cGoSetEventCallback(void* clockCtx) {
		// The 'globalEventCallback' Go function is shared amongst all possible Clock instances.

		// Given that the 'globalEventCallback' is shared, we pass again the
		// clockCtx instance but in this case is needed to pick up the correct method
		// that will handle the event.

		// In other words, for every call libclock makes to globalEventCallback,
		// the 'userData' parameter will bring the context of the clock that registered
		// that globalEventCallback.

		// This technique is needed because cgo only allows to export Go functions and not methods.

		clock_set_event_callback(clockCtx, (ClockCallBack) globalEventCallback, clockCtx);
	}

	static void cGoClockDestroy(void* clockCtx, void* resp) {
		clock_destroy(clockCtx, (ClockCallBack) GoCallback, resp);
	}

	static void cGoClockSetAlarm(void* clockCtx, int timeMillis, const char* alarmMsg, void* resp) {
		clock_set_alarm(clockCtx, timeMillis, alarmMsg, (ClockCallBack) GoCallback, resp);
	}

	static void cGoListAlarms(void* clockCtx, void* resp) {
		clock_list_alarms(clockCtx, (ClockCallBack) GoCallback, resp);
	}

*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
	"unsafe"
)

const requestTimeout = 30 * time.Second
const EventChanBufferSize = 1024

//export GoCallback
func GoCallback(ret C.int, msg *C.char, len C.size_t, resp unsafe.Pointer) {
	if resp != nil {
		m := (*C.Resp)(resp)
		m.ret = ret
		m.msg = msg
		m.len = len
		wg := (*sync.WaitGroup)(m.ffiWg)
		wg.Done()
	}
}

type EventCallbacks struct {
	OnAlarm func(time string, msg string)
}

// Clock represents an instance of a nim-c-library-guide Clock
type Clock struct {
	clockCtx  unsafe.Pointer
	callbacks EventCallbacks
}

func NewClock() (*Clock, error) {
	Debug("Creating new Clock")
	clock := &Clock{}

	wg := sync.WaitGroup{}

	var resp = C.allocResp(unsafe.Pointer(&wg))

	defer C.freeResp(resp)

	if C.getRet(resp) != C.RET_OK {
		errMsg := C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
		Error("error NewClock: %v", errMsg)
		return nil, errors.New(errMsg)
	}

	wg.Add(1)
	clock.clockCtx = C.cGoNewClock(resp)
	wg.Wait()

	C.cGoSetEventCallback(clock.clockCtx)
	registerClock(clock)

	Debug("Successfully created Clock")
	return clock, nil
}

// The event callback sends back the clock ctx to know to which
// clock is the event being emited for. Since we only have a global
// callback in the go side, We register all the clock's that we create
// so we can later obtain which instance of `Clock` it should
// be invoked depending on the ctx received

var clockRegistry map[unsafe.Pointer]*Clock

func init() {
	clockRegistry = make(map[unsafe.Pointer]*Clock)
}

func registerClock(clock *Clock) {
	_, ok := clockRegistry[clock.clockCtx]
	if !ok {
		clockRegistry[clock.clockCtx] = clock
	}
}

func unregisterClock(clock *Clock) {
	delete(clockRegistry, clock.clockCtx)
}

//export globalEventCallback
func globalEventCallback(callerRet C.int, msg *C.char, len C.size_t, userData unsafe.Pointer) {
	if callerRet == C.RET_OK {
		eventStr := C.GoStringN(msg, C.int(len))
		clock, ok := clockRegistry[userData] // userData contains clock's ctx
		if ok {
			clock.OnEvent(eventStr)
		}
	} else {
		if len != 0 {
			errMsg := C.GoStringN(msg, C.int(len))
			Error("globalEventCallback retCode not ok, retCode: %v: %v", callerRet, errMsg)
		} else {
			Error("globalEventCallback retCode not ok, retCode: %v", callerRet)
		}
	}
}

type jsonEvent struct {
	EventType string `json:"eventType"`
}

type alarmEvent struct {
	Time string `json:"time"`
	Msg  string `json:"msg"`
}

func (c *Clock) RegisterCallbacks(callbacks EventCallbacks) {
	c.callbacks = callbacks
}

func (c *Clock) OnEvent(eventStr string) {

	jsonEvent := jsonEvent{}
	err := json.Unmarshal([]byte(eventStr), &jsonEvent)
	if err != nil {
		Error("could not unmarshal event string: %v", err)

		return
	}

	switch jsonEvent.EventType {
	case "clock_alarm":
		c.parseAlarmEvent(eventStr)

	}

}

func (c *Clock) parseAlarmEvent(eventStr string) {

	alarmEvent := alarmEvent{}
	err := json.Unmarshal([]byte(eventStr), &alarmEvent)
	if err != nil {
		Error("could not parse alarm event %v", err)
	}

	if c.callbacks.OnAlarm != nil {
		c.callbacks.OnAlarm(alarmEvent.Time, alarmEvent.Msg)
	}
}

func (c *Clock) Destroy() error {
	if c == nil {
		err := errors.New("clock is nil")
		Error("Failed to destroy %v", err)
		return err
	}

	Debug("Destroying clock")

	wg := sync.WaitGroup{}
	var resp = C.allocResp(unsafe.Pointer(&wg))
	defer C.freeResp(resp)

	wg.Add(1)
	C.cGoClockDestroy(c.clockCtx, resp)
	wg.Wait()

	if C.getRet(resp) == C.RET_OK {
		unregisterClock(c)
		Debug("Successfully destroyed clock")
		return nil
	}

	errMsg := "error Destroy: " + C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	Error("Failed to destroy clock: %v", errMsg)

	return errors.New(errMsg)
}

func (c *Clock) SetAlarm(timeMillis int, alarmMsg string) error {
	Debug("Setting alarm in %v millis", timeMillis)

	wg := sync.WaitGroup{}

	var resp = C.allocResp(unsafe.Pointer(&wg))
	var cAlarmMsg = C.CString(string(alarmMsg))
	defer C.freeResp(resp)
	defer C.free(unsafe.Pointer(cAlarmMsg))

	wg.Add(1)
	C.cGoClockSetAlarm(c.clockCtx, C.int(timeMillis), cAlarmMsg, resp)
	wg.Wait()
	if C.getRet(resp) == C.RET_OK {
		Debug("Successfully set alarm in %v millis", timeMillis)
		return nil
	}
	errMsg := "error SetAlarm: " +
		C.GoStringN(C.getMyCharPtr(resp), C.int(C.getMyCharLen(resp)))
	return fmt.Errorf("SetAlarm: %s", errMsg)
}
