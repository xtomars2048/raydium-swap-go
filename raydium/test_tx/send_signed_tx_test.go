package test_tx

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/katelouis/raydium-swap-go/raydium/test_tx/types"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestSendSignedTx(t *testing.T) {

	var req types.ReqBroadcastTx
	req.RequestId = strconv.FormatInt(time.Now().Unix(), 10)
	req.Signatures = make([][64]byte, 0)
	req.TxRawMessage = ""

	connection := rpc.New(os.Getenv("RPC_URL"))

	var transaction solana.Transaction
	transaction.Message.UnmarshalBase64(req.TxRawMessage)
	transaction.Signatures = []solana.Signature(req.Signatures[:])

	signature, err := connection.SendTransactionWithOpts(context.Background(), &transaction, rpc.TransactionOpts{SkipPreflight: true})

	if err != nil {
		panic(err)
	}
	fmt.Println("Transaction successfully sent: ", signature)
}
