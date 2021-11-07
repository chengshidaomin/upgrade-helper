package testdata

import (
	"fmt"
	"net/http"
)

func data2() {
	fmt.Println(http.SameSiteDefaultMode)
}
