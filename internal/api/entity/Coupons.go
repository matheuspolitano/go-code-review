package entity

type CouponRequest struct {
	Discount       int    `json:"discount"`
	Code           string `json:"code"`
	MinBasketValue int    `json:"min_basket_value"`
}
