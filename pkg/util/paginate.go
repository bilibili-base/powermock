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
