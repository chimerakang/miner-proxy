package config

type Proxy struct {
	PoolFee        *float32 `json:"poolFee"`
	PoolFeeAddress *string  `json:"poolFeeAddress"`
}
