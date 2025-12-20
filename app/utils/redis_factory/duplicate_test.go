package redis_factory

import (
	"fmt"
	"testing"
)

func TestIsDuplicateRequest(t *testing.T) {
	f := NewDuplicateFactory("kaka", 50)
	err := f.IsDuplicateRequest()
	fmt.Printf("IsDuplicateRequest err:%v", err)
	err = f.Clean()
	fmt.Printf("Clean err:%v", err)
}
