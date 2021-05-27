// Copyright 2021 bilibili-base
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

package util

import (
	"fmt"
	"reflect"

	"github.com/bilibili-base/powermock/apis/v1alpha1"
)

// GetPagination is used to make ListOptions safety
func GetPagination(options *v1alpha1.ListOptions) *v1alpha1.ListOptions {
	if options == nil {
		options = &v1alpha1.ListOptions{}
	}
	if options.Page == 0 {
		options.Page = 1
	}
	if options.Limit == 0 {
		options.Limit = 10
	}
	return options
}

// PaginateSlice is used to paginate slice
func PaginateSlice(options *v1alpha1.ListOptions, pointer interface{}) error {
	v := reflect.ValueOf(pointer)
	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("non-pointer %v", v.Type())
	}
	// get the value that the pointer v points to.
	v = v.Elem()
	if v.Kind() != reflect.Slice {
		return fmt.Errorf("can't fill non-slice value")
	}

	page := options.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * options.Limit
	limit := options.Limit
	count := uint64(v.Len())
	if offset >= count {
		offset = 0
		limit = 0
	}
	if offset+limit > count {
		limit = count - offset
	}
	v.Set(v.Slice(int(offset), int(offset+limit)))
	return nil
}
