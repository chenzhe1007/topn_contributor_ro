package fetcher

import (
	"fmt"
	"testing"

	"github.com/tomnomnom/linkheader"
)

func TestLinkParse(t *testing.T) {
	links := "<https://api.github.com/user/3299148/repos?page=2>; rel=\"next\", <https://api.github.com/user/3299148/repos?page=17>; rel=\"last\"] Referrer-Policy:[origin-when-cross-origin, strict-origin-when-cross-origin"
	for _, link := range linkheader.Parse(links) {
		fmt.Println(link.Param("page"))
	}
}
