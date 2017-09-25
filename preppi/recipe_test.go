// Copyright (c) 2017 Christian Funkhouser <christian.funkhouser@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package preppi

import (
	"reflect"
	"testing"
)

func TestRecipeVars(t *testing.T) {
	r := &Recipe{
		Name: "Test Recipe",
		Ingredients: []*Ingredient{
			&Ingredient{Vars: []string{"Foo", "Bar", "Baz"}},
			&Ingredient{Vars: []string{"Foo", "Bar", "Quux"}},
		},
	}

	want := []string{"Bar", "Baz", "Foo", "Quux"}
	got := r.Vars()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("wanted %+v, got %+v", want, got)
	}
}

func TestIngredientCompileTemplate(t *testing.T) {
	for _, tt := range []struct {
		tmplString string
		wantErr    bool
	}{
		{
			tmplString: `{{.Something}}`,
			wantErr:    false,
		}, {
			tmplString: `{{.Something`,
			wantErr:    true,
		},
	} {
		ingredient := &Ingredient{
			Source: "/some/file/path",
		}
		_, err := ingredient.compileTemplate(tt.tmplString)
		if err != nil && !tt.wantErr {
			t.Errorf("wanted no error, got: %v", err)
		} else if err == nil && tt.wantErr {
			t.Error("wanted error, but got none")
		}
	}
}
