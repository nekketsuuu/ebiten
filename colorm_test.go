// Copyright 2014 Hajime Hoshi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ebiten_test

import (
	. "github.com/hajimehoshi/ebiten"
	"testing"
)

func TestColorInit(t *testing.T) {
	var m ColorM
	for i := 0; i < ColorMDim-1; i++ {
		for j := 0; j < ColorMDim; j++ {
			got := m.Element(i, j)
			want := 0.0
			if i == j {
				want = 1
			}
			if want != got {
				t.Errorf("m.Element(%d, %d) = %f, want %f", i, j, got, want)
			}
		}
	}

	m.SetElement(0, 0, 1)
	for i := 0; i < ColorMDim-1; i++ {
		for j := 0; j < ColorMDim; j++ {
			got := m.Element(i, j)
			want := 0.0
			if i == j {
				want = 1
			}
			if want != got {
				t.Errorf("m.Element(%d, %d) = %f, want %f", i, j, got, want)
			}
		}
	}
}

func TestColorAssign(t *testing.T) {
	m := ColorM{}
	m.SetElement(0, 0, 1)
	m2 := m
	m.SetElement(0, 0, 0)
	got := m2.Element(0, 0)
	want := 1.0
	if want != got {
		t.Errorf("m2.Element(%d, %d) = %f, want %f", 0, 0, got, want)
	}
}

func TestColorTranslate(t *testing.T) {
	expected := [4][5]float64{
		{1, 0, 0, 0, 0.5},
		{0, 1, 0, 0, 1.5},
		{0, 0, 1, 0, 2.5},
		{0, 0, 0, 1, 3.5},
	}
	m := ColorM{}
	m.Translate(0.5, 1.5, 2.5, 3.5)
	for i := 0; i < 4; i++ {
		for j := 0; j < 5; j++ {
			got := m.Element(i, j)
			want := expected[i][j]
			if want != got {
				t.Errorf("m.Element(%d, %d) = %f, want %f", i, j, got, want)
			}
		}
	}
}

func TestColorScale(t *testing.T) {
	expected := [4][5]float64{
		{0.5, 0, 0, 0, 0},
		{0, 1.5, 0, 0, 0},
		{0, 0, 2.5, 0, 0},
		{0, 0, 0, 3.5, 0},
	}
	m := ColorM{}
	m.Scale(0.5, 1.5, 2.5, 3.5)
	for i := 0; i < 4; i++ {
		for j := 0; j < 5; j++ {
			got := m.Element(i, j)
			want := expected[i][j]
			if want != got {
				t.Errorf("m.Element(%d, %d) = %f, want %f", i, j, got, want)
			}
		}
	}
}

func TestColorTranslateAndScale(t *testing.T) {
	expected := [4][5]float64{
		{1, 0, 0, 0, 0},
		{0, 1, 0, 0, 0},
		{0, 0, 1, 0, 0},
		{0, 0, 0, 0.5, 0.5},
	}
	m := ColorM{}
	m.Translate(0, 0, 0, 1)
	m.Scale(1, 1, 1, 0.5)
	for i := 0; i < 4; i++ {
		for j := 0; j < 5; j++ {
			got := m.Element(i, j)
			want := expected[i][j]
			if want != got {
				t.Errorf("m.Element(%d, %d) = %f, want %f", i, j, got, want)
			}
		}
	}
}

func TestColorMonochrome(t *testing.T) {
	expected := [4][5]float64{
		{0.2990, 0.5870, 0.1140, 0, 0},
		{0.2990, 0.5870, 0.1140, 0, 0},
		{0.2990, 0.5870, 0.1140, 0, 0},
		{0, 0, 0, 1, 0},
	}
	m := Monochrome()
	for i := 0; i < 4; i++ {
		for j := 0; j < 5; j++ {
			got := m.Element(i, j)
			want := expected[i][j]
			if want != got {
				t.Errorf("m.Element(%d, %d) = %f, want %f", i, j, got, want)
			}
		}
	}
}
