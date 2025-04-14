package types

type CommonResponse struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
}

type ReqBroadcastTx struct {
	Signatures [][64]byte `json:"signatures"`
	TxMessage  string     `json:"txMessage"`
}

type ReqBuildTx struct {
	FromAddress        string `json:"fromAddress"`
	InputToken         string `json:"inputToken"`
	InputTokenDecimal  int32  `json:"inputTokenDecimal"`
	OutputToken        string `json:"outputToken"`
	OutputTokenDecimal int32  `json:"outputTokenDecimal"`
	Slippage           string `json:"slippage"`
	Amount             string `json:"amount"`
	Fee                uint64 `json:"fee"`
}

type RespBroadcastTx struct {
	CommonResponse
	Data string `json:"data"`
}

type RespBuildTx struct {
	CommonResponse
	Data string `json:"data"`
}
