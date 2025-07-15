package models

type PaginationInfo struct {
	CurrentPage  int
	TotalPages   int
	TotalRecords int
	PageSize     int
	Pages        []PageInfo
}

type PageInfo struct {
	Number     int
	IsCurrent  bool
	IsEllipsis bool
}

func (p *PaginationInfo) OrdinalNumber(index int) int {
	return (p.CurrentPage-1)*p.PageSize + index + 1
}
