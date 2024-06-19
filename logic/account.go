package logic

type Account struct {
	ID int32 `json:"ID"`

	AccountName string `json:"AccountName"`

	DisplayName string `json:"DisplayName"`

	Role string `json:"Role"`
}
