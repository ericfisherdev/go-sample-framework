package session

import (
	"fmt"
	"github.com/alexedwards/scs/v2"
	"reflect"
	"testing"
)

func TestSession_InitSession(t *testing.T) {

	f := &Session{
		CookieLifetime: "100",
		CookiePersist:  "true",
		CookieName:     "finishline",
		CookieDomain:   "localhost",
		SessionType:    "cookie",
	}

	var sm *scs.SessionManager

	ses := f.InitSession()

	var sessKind reflect.Kind
	var sessType reflect.Type

	rv := reflect.ValueOf(ses)

	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		fmt.Println("For loop:", rv.Kind(), rv.Type(), rv)
		sessKind = rv.Kind()
		sessType = rv.Type()

		rv = rv.Elem()
	}

	if !rv.IsValid() {
		t.Error("Invalid type or kind; kind:", rv.Kind(), "type:", rv.Type())
	}

	if sessKind != reflect.ValueOf(sm).Kind() {
		t.Error("Wrong kind returned testing cookie session. Expected", reflect.ValueOf(sm).Kind(), "and got", sessKind)
	}

	if sessType != reflect.ValueOf(sm).Type() {
		t.Error("Wrong type returned testing cookie session. Expected", reflect.ValueOf(sm).Type(), "and got", sessType)
	}
}
