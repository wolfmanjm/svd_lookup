package svd2db

import (
	"testing"
)

func TestConvert(t *testing.T) {
	fn := "testdata/test.svd"
	err := Convert(fn, "")
	if err != nil {
        t.Errorf(`Convert("%v") = %v, want nil`, fn, err)
    }
}
