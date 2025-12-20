package data_type

type Page struct {
	Page  float64 `form:"page" json:"page"`   // 必填，页面值>=1
	Limit float64 `form:"limit" json:"limit"` // 必填，每页条数值>=1
}

func (p *Page) SetDefault() {
	if p != nil {
		if p.Page <= 0 {
			p.Page = 1
		}
		if p.Limit <= 0 {
			p.Limit = 20
		}
	}
}
