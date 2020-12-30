// Copyright 2020 The LUCI Authors.
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

package common

import (
	"sort"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestIDs(t *testing.T) {
	t.Parallel()

	Convey("IDs helper works", t, func() {
		So(MakeRunIDs().WithoutSorted(MakeRunIDs("1")), ShouldResemble, MakeRunIDs())
		So(RunIDs(nil).WithoutSorted(MakeRunIDs("1")), ShouldEqual, nil)

		ids := MakeRunIDs("5", "8", "2")
		sort.Sort(ids)
		So(ids, ShouldResemble, MakeRunIDs("2", "5", "8"))

		So(ids.Equal(MakeRunIDs("2", "5", "8")), ShouldBeTrue)
		So(ids.Equal(MakeRunIDs("2", "5", "8", "8")), ShouldBeFalse)

		assertSameSlice(ids.WithoutSorted(nil), ids)
		So(ids, ShouldResemble, MakeRunIDs("2", "5", "8"))

		assertSameSlice(ids.WithoutSorted(MakeRunIDs("1", "3", "9")), ids)
		So(ids, ShouldResemble, MakeRunIDs("2", "5", "8"))

		So(ids.WithoutSorted(MakeRunIDs("1", "5", "9")), ShouldResemble, MakeRunIDs("2", "8"))
		So(ids, ShouldResemble, MakeRunIDs("2", "5", "8"))

		So(ids.WithoutSorted(MakeRunIDs("1", "5", "5", "7")), ShouldResemble, MakeRunIDs("2", "8"))
		So(ids, ShouldResemble, MakeRunIDs("2", "5", "8"))
	})
}

func assertSameSlice(a, b RunIDs) {
	// Go doesn't allow comparing slices, so compare their contents and ensure
	// pointers to the first element are the same.
	So(a, ShouldResemble, b)
	So(&a[0], ShouldEqual, &b[0])
}