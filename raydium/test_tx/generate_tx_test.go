package test_tx

import (
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/katelouis/raydium-swap-go/raydium"
	"github.com/katelouis/raydium-swap-go/raydium/test_tx/types"
	"github.com/katelouis/raydium-swap-go/raydium/trade"
	"github.com/katelouis/raydium-swap-go/raydium/utils"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestGenerateTx(t *testing.T) {

	var req types.ReqBuildTx
	req.RequestId = strconv.FormatInt(time.Now().Unix(), 10)
	req.FromAddress = "CeA4SAdvZF8dEoQWnUQvN1VG9fUZSVX6DGhgpHMPBwic"
	req.InputToken = "So11111111111111111111111111111111111111112"
	req.InputTokenDecimal = 9
	req.OutputToken = "6p6xgHyF7AeE6TZkSmFsko444wqoP15icUSqi2jfGiPN"
	req.OutputTokenDecimal = 6
	req.Slippage = "0.01"
	req.Amount = "0.001"
	req.Fee = 25000

	connection := rpc.New(os.Getenv("RPC_URL"))
	raydium := raydium.New(connection, req.FromAddress)

	inputToken := utils.NewToken("", req.InputToken, uint64(req.InputTokenDecimal))
	outputToken := utils.NewToken("", req.OutputToken, uint64(req.OutputTokenDecimal))

	fSlip, err := strconv.ParseFloat(req.Slippage, 64)
	if err != nil {
		logrus.Warnln("ParseFloat(req.Slippage, 64)  error", err, "req:", req)
		return
	}

	// slippage := utils.NewPercent(1, 100) // 1% slippage (for 0.5 set second parameter to "1000" example: utils.NewPercent(5, 1000) )
	slippage := utils.NewPercent(uint64(fSlip*1000), 1000)

	fAmount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil {
		logrus.Warnln("ParseFloat(req.Amount, 64)  error", err, "req:", req)
		return
	}
	inputAmount := utils.NewTokenAmount(inputToken, fAmount)

	poolKeys, err := raydium.Pool.GetPoolKeys(req.InputToken, outputToken.Mint)
	if err != nil {
		logrus.Warnln("GetPoolKeys  error", err, "req:", req)
		return
	}

	amountsOut, err := raydium.Liquidity.GetAmountsOut(poolKeys, inputAmount, slippage)
	if err != nil {
		logrus.Warnln("GetAmountsOut  error", err, "req:", req)
		return
	}
	FeeVault := os.Getenv("Vault")

	tx, err := raydium.Trade.MakeRawSwapTx(
		poolKeys,
		amountsOut.AmountIn,
		amountsOut.MinAmountOut,
		trade.FeeConfig{
			MicroLamports: req.Fee, // fee 0.000025 sol
		},
		FeeVault,
		10000,
	)
	if err != nil {
		logrus.Warnln("MakeRawSwapTx  error", err, "req:", req)
		return
	}

	logrus.Infof("kelly output:>>> " + tx.Message.ToBase64() + "\n")
	SendSignedTx(tx.Message.ToBase64())
}
