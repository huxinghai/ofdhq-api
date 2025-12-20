package tronscanapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"ofdhq-api/app/global/variable"
	"strconv"
	"time"

	"github.com/qifengzhang007/goCurl"
	"go.uber.org/zap"
)

func GetTransfers(toAddress string, startTime, endTime *time.Time, start, limit int64) (*TransferWrap, error) {
	args := map[string]string{
		"toAddress":        toAddress,
		"filterTokenValue": "1",
		"limit":            strconv.FormatInt(limit, 10),
		"start":            strconv.FormatInt(start, 10),
	}

	if startTime != nil && endTime != nil {
		startMilliseconds := startTime.UnixNano() / int64(time.Millisecond)
		endMilliseconds := endTime.UnixNano() / int64(time.Millisecond)
		args["start_timestamp"] = fmt.Sprint(startMilliseconds)
		args["end_timestamp"] = fmt.Sprint(endMilliseconds)
	}

	queryParams := url.Values{}
	for key, value := range args {
		queryParams.Add(key, value)
	}
	argString := queryParams.Encode()
	variable.ZapLog.Info("请求GetTransfers req", zap.String("argString", argString))

	cli := goCurl.CreateHttpClient(goCurl.Options{
		Headers: map[string]interface{}{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		SetResCharset: "utf-8",
	})

	resp, err := cli.Get("https://apilist.tronscanapi.com/api/token_trc20/transfers?confirm=true&" + argString)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("GetTransfers 请求报错"))
	}
	body, err := resp.GetContents()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("GetTransfers 获取Body报错"))
	}
	result := &TransferWrap{}
	err = json.Unmarshal([]byte(body), result)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("GetTransfers json.Unmarshal body:%s", body))
	}

	return result, nil
}

type Transfer struct {
	TransactionID   string           `json:"transaction_id"`
	Status          int              `json:"status"`
	BlockTs         int64            `json:"block_ts"`
	FromAddress     string           `json:"from_address"`
	FromAddressTag  *json.RawMessage `json:"from_address_tag"`
	ToAddress       string           `json:"to_address"`
	ToAddressTag    *json.RawMessage `json:"to_address_tag"`
	Block           int64            `json:"block"`
	ContractAddress string           `json:"contract_address"`
	Quant           string           `json:"quant"`
	Confirmed       bool             `json:"confirmed"`
	ContractRet     string           `json:"contractRet"`
	FinalResult     string           `json:"finalResult"`
	Revert          bool             `json:"revert"`
	TokenInfo       struct {
		TokenID      string `json:"tokenId"`
		TokenAbbr    string `json:"tokenAbbr"`
		TokenName    string `json:"tokenName"`
		TokenDecimal int64  `json:"tokenDecimal"`
		TokenCanShow int64  `json:"tokenCanShow"`
		TokenType    string `json:"tokenType"`
		TokenLogo    string `json:"tokenLogo"`
		TokenLevel   string `json:"tokenLevel"`
		IssuerAddr   string `json:"issuerAddr"`
		Vip          bool   `json:"vip"`
	} `json:"tokenInfo"`
	ContractType          string `json:"contract_type"`
	FromAddressIsContract bool   `json:"fromAddressIsContract"`
	ToAddressIsContract   bool   `json:"toAddressIsContract"`
	RiskTransaction       bool   `json:"riskTransaction"`
}

type TransferWrap struct {
	Total             int64            `json:"total"`
	RangeTotal        int64            `json:"rangeTotal"`
	ContractInfo      *json.RawMessage `json:"contractInfo"`
	TokenTransfers    []*Transfer      `json:"token_transfers"`
	TimeInterval      int64            `json:"timeInterval"`
	NormalAddressInfo *json.RawMessage `json:"normalAddressInfo"`
	Message           string           `json:"message"`
}
