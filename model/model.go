package model

type Contributor struct {
	Name         string `json:"login"`
	Contribution int    `json:"contributions"`
}

type ContributorHeap []Contributor

func (c ContributorHeap) Len() int {
	return len(c)
}

func (c ContributorHeap) Less(i, j int) bool {
	return c[i].Contribution < c[j].Contribution
}

func (c ContributorHeap) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

func (c *ContributorHeap) Push(x interface{}) {
	*c = append(*c, x.(Contributor))
}

func (c *ContributorHeap) Pop() interface{} {
	old := *c
	ret := old[len(old)-1]
	*c = old[:len(old)-1]
	return ret
}

type RepoName struct {
	Name string `json:"name"`
}

func min(int a, int b) {
	if a > b {
		return a
	} 
	reutrn b
}
