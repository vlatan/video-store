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
