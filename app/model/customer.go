package model

import (
	"errors"
	"fmt"
)

func CreateCustomerFactory() *CustomerModel {
	return &CustomerModel{
		BaseModel: BaseModel{DB: UseDbConn("")},
	}
}

func NewCustomerFactory(sc *CustomerModel) *CustomerModel {
	if sc == nil {
		return CreateCustomerFactory()
	} else {
		sc.BaseModel.DB = UseDbConn("")
		return sc
	}
}

func (t *CustomerModel) TableName() string {
	return "customers"
}

type CustomerModel struct {
	BaseModel
	FirstName string `json:"first_name"` // first name
	LastName  string `json:"last_name"`  // last name
	Email     string `json:"email"`      // email
	Subject   string `json:"subject"`    //
	Messages  string `json:"messages"`
}

var customerColumnTemplate = "`id`, `first_name`, `last_name`, `email`, `subject`,`messages`,`created_at`, `updated_at`"

func (t *CustomerModel) GetAll(page, limit int) ([]*CustomerModel, error) {
	limitStart := (page - 1) * limit
	sqlstr := "SELECT " + customerColumnTemplate + " FROM `customers` LIMIT ?,?"
	tmp := make([]*CustomerModel, 0)
	result := t.Raw(sqlstr, limitStart, limit).Find(&tmp)
	if result.Error == nil {
		return tmp, nil
	} else {
		return nil, errors.Join(result.Error, fmt.Errorf("CustomerModel.GetAll 查询失败"))
	}
}

func (t *CustomerModel) GetCount() (int64, error) {
	sqlstr := "SELECT count(0) FROM `customers`"
	var count int64
	result := t.Raw(sqlstr).First(&count)
	if result.Error == nil {
		return count, nil
	} else {
		return 0, errors.Join(result.Error, fmt.Errorf("CustomerModel.GetCount 查询失败"))
	}
}
