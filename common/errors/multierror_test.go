// Copyright 2015 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestMultiError(t *testing.T) {
	t.Parallel()

	Convey("MultiError works", t, func() {
		var me error = MultiError{fmt.Errorf("hello"), fmt.Errorf("bob")}

		So(me.Error(), ShouldEqual, `hello (and 1 other error)`)
	})
}

func TestUpstreamErrors(t *testing.T) {
	t.Parallel()

	Convey("Test MultiError", t, func() {
		Convey("nil", func() {
			me := MultiError(nil)
			So(me.Error(), ShouldEqual, "(0 errors)")
			Convey("single", func() {
				So(SingleError(error(me)), ShouldBeNil)
			})
		})
		Convey("one", func() {
			me := MultiError{errors.New("sup")}
			So(me.Error(), ShouldEqual, "sup")
		})
		Convey("two", func() {
			me := MultiError{errors.New("sup"), errors.New("what")}
			So(me.Error(), ShouldEqual, "sup (and 1 other error)")
		})
		Convey("more", func() {
			me := MultiError{errors.New("sup"), errors.New("what"), errors.New("nerds")}
			So(me.Error(), ShouldEqual, "sup (and 2 other errors)")

			Convey("single", func() {
				So(SingleError(error(me)), ShouldResemble, errors.New("sup"))
			})
		})
	})

	Convey("MaybeAdd", t, func() {
		me := MultiError(nil)

		Convey("nil", func() {
			me.MaybeAdd(nil)
			So(me, ShouldHaveLength, 0)
			So(error(me), ShouldBeNil)
		})

		Convey("thing", func() {
			me.MaybeAdd(errors.New("sup"))
			So(me, ShouldHaveLength, 1)
			So(error(me), ShouldNotBeNil)

			me.MaybeAdd(errors.New("what"))
			So(me, ShouldHaveLength, 2)
			So(error(me), ShouldNotBeNil)
		})

	})

	Convey("AsError", t, func() {
		var me MultiError
		So(me, ShouldBeNil)

		var err error
		err = me

		// Unfortunately Go has many nil's :(
		//   So(err == nil, ShouldBeTrue)
		// Note that `ShouldBeNil` won't cut it, since it 'sees through' interfaces.

		// However!
		err = me.AsError()
		So(err == nil, ShouldBeTrue)
	})

	Convey("SingleError passes through", t, func() {
		e := errors.New("unique")
		So(SingleError(e), ShouldEqual, e)
	})
}

func TestFlatten(t *testing.T) {
	t.Parallel()

	Convey("Flatten works", t, func() {
		Convey("Nil", func() {
			So(Flatten(MultiError{nil, nil, MultiError{nil, nil, nil}}), ShouldBeNil)
		})

		Convey("2-dim", func() {
			So(Flatten(MultiError{nil, errors.New("1"), nil, MultiError{nil, errors.New("2"), nil}}),
				ShouldResemble, MultiError{errors.New("1"), errors.New("2")})
		})

		Convey("Doesn't unwrap", func() {
			ann := Annotate(MultiError{nil, nil, nil}, "don't do this").Err()
			merr, yup := Flatten(MultiError{nil, ann, nil, MultiError{nil, errors.New("2"), nil}}).(MultiError)
			So(yup, ShouldBeTrue)
			So(len(merr), ShouldEqual, 2)
			So(merr, ShouldResemble, MultiError{ann, errors.New("2")})
		})
	})
}
