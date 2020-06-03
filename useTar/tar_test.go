package useTar

import (
	"fmt"
	"testing"
)

func TestTarGz(t *testing.T) {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	path := "/home/wjk/go/studyProject/base/factory"
	name := "factory.tar.gz"

	err := TarGz(path, name, 0)
	if err != nil {
		panic(err)
	}

}
