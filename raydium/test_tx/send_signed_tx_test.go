package test_tx

import (
	"context"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/katelouis/raydium-swap-go/raydium/test_tx/types"
	"github.com/mr-tron/base58"
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
	"testing"
	"time"
)

func TestSendSignedTx(t *testing.T) {

	var req types.ReqBroadcastTx
	req.RequestId = strconv.FormatInt(time.Now().Unix(), 10)
	req.TxRawMessage = "AQAJFd6Y1SYwImshDdlJqszTOd8FCjuoxLeRrFeJeidom3evJP3MggN1CuqD6P7s7vf2+k9sCFYZApMRHd1SoB2m+S7zfJQyQ1zrOhUkOxm8BmFPg8d4k/P2OQBHEs+rFYcynhUdMiZtm2RUG5dbV4WfKGJqQnDIpLNqFM235emApkot/A9horYULnL0VE7oZrSeOZ2iRIqSQBmbhWiWt0KtnOBjJyUw3wLRBtOMYC8M4pi+I64hMsM964cOCcWbFlT4qoY3QNeYqHMiO8DSdhsEiSCOrTebQQ3YoqgrMQBrvORlw1zybvCy5wGnASoXo7h0tLftbvh8knNXy1QMdxFe9Vn3DUzOGZ0lBNTyZ1OR0TAzNBSM7XMuWon1mh+N0IWhNNaMgh32eyxkgcZQ2KysPcLsc14zDHLceU2SpNl2itO1okmi+3vkRQrDpP77yXrg1Si0StapLMCSYaCE2wYa6gDgnx7vQMCTl994QVX/SWOq9ONpLDodvWTbid9kFCjW4wabiFf+q4GE+2h/Y0YYwDXaxDncGus7VZig8AAAAAABBqfVFxksXFEhjMlMPUrxf1ja7gibof1E49vZigAAAAAG3fbh12Whk9nL4UbO63msHLSF7V9bN5E6jPWFfv8AqUFXsFgPMcX85EpiWC28+deO51lDoISjk7NQNo0iiZMIDQdRqCgtphMF/imcN7mY5YRx2xE1A3MQ+L4QRaYK9u5mnhqRQIAMtKZEjkHy6Mrx4UZJlUpoe/8TZtwda+71iwMGRm/lIRcy/+ytunLDm+e8jOW7xfcSayxDmzpAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABL2UnENgLDPyB3kO0Wo1JMobmXXPEhoqkM/+x9+LaKzfFnDr99oajqx6SO320hZQNe+bbiF48rWGNlSotLEV5IBhIABQLAJwkAEgAJA6hhAAAAAAAAEwMAAQB8AwAAAN6Y1SYwImshDdlJqszTOd8FCjuoxLeRrFeJeidom3evIAAAAAAAAABEUHI0ekZUUkE1dnplUnhDOEVtRTQ0emN4bmZFeG1QYZCkIAAAAAAApQAAAAAAAAAG3fbh12Whk9nL4UbO63msHLSF7V9bN5E6jPWFfv8AqQ4EAQwADQEBFBIOAg8DBAUGEAcICQoFBhEBCwARCaCGAQAAAAAARAYAAAAAAAAOAwEAAAEJ"

	connection := rpc.New(os.Getenv("RPC_URL"))

	// ------------------------------------------------
	var txTemp solana.Transaction
	txTemp.Message.UnmarshalBase64(req.TxRawMessage)

	privateKey := "4To1e8EAJdgngKF8ts7AXWozemEth2SjqsmmbBesnE4NUQw32JSxfuiGLMEs2jvFFLYjahj6WtHsaxp5u8rMVzjg"
	pK, err := base58.Decode(privateKey)

	pkPrivateKey := solana.PrivateKey(pK)

	_, err = txTemp.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &pkPrivateKey
	})
	if err != nil {

	}
	req.Signatures = make([][64]byte, len(txTemp.Signatures))

	for i, sig := range txTemp.Signatures {
		req.Signatures[i] = sig
	}

	// ------------------------------------------------

	var transaction solana.Transaction
	transaction.Message.UnmarshalBase64(req.TxRawMessage)

	transaction.Signatures = make([]solana.Signature, len(req.Signatures))
	for i, sig := range req.Signatures {
		transaction.Signatures[i] = solana.Signature(sig)
	}

	signature, err := connection.SendTransactionWithOpts(context.Background(), &transaction, rpc.TransactionOpts{SkipPreflight: true})

	if err != nil {
		panic(err)
	}
	fmt.Println("Transaction successfully sent: ", signature)
}

func SendSignedTx(TxRawMessage string) {

	var req types.ReqBroadcastTx
	req.RequestId = strconv.FormatInt(time.Now().Unix(), 10)
	// req.TxRawMessage = "AQAJFd6Y1SYwImshDdlJqszTOd8FCjuoxLeRrFeJeidom3evJP3MggN1CuqD6P7s7vf2+k9sCFYZApMRHd1SoB2m+S7zfJQyQ1zrOhUkOxm8BmFPg8d4k/P2OQBHEs+rFYcynhUdMiZtm2RUG5dbV4WfKGJqQnDIpLNqFM235emApkot/A9horYULnL0VE7oZrSeOZ2iRIqSQBmbhWiWt0KtnOBjJyUw3wLRBtOMYC8M4pi+I64hMsM964cOCcWbFlT4qoY3QNeYqHMiO8DSdhsEiSCOrTebQQ3YoqgrMQBrvORlw1zybvCy5wGnASoXo7h0tLftbvh8knNXy1QMdxFe9Vn3DUzOGZ0lBNTyZ1OR0TAzNBSM7XMuWon1mh+N0IWhNNaMgh32eyxkgcZQ2KysPcLsc14zDHLceU2SpNl2itO1okmi+3vkRQrDpP77yXrg1Si0StapLMCSYaCE2wYa6gDgnx7vQMCTl994QVX/SWOq9ONpLDodvWTbid9kFCjW4wabiFf+q4GE+2h/Y0YYwDXaxDncGus7VZig8AAAAAABBqfVFxksXFEhjMlMPUrxf1ja7gibof1E49vZigAAAAAG3fbh12Whk9nL4UbO63msHLSF7V9bN5E6jPWFfv8AqUFXsFgPMcX85EpiWC28+deO51lDoISjk7NQNo0iiZMIDQdRqCgtphMF/imcN7mY5YRx2xE1A3MQ+L4QRaYK9u5mnhqRQIAMtKZEjkHy6Mrx4UZJlUpoe/8TZtwda+71iwMGRm/lIRcy/+ytunLDm+e8jOW7xfcSayxDmzpAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABL2UnENgLDPyB3kO0Wo1JMobmXXPEhoqkM/+x9+LaKzfFnDr99oajqx6SO320hZQNe+bbiF48rWGNlSotLEV5IBhIABQLAJwkAEgAJA6hhAAAAAAAAEwMAAQB8AwAAAN6Y1SYwImshDdlJqszTOd8FCjuoxLeRrFeJeidom3evIAAAAAAAAABEUHI0ekZUUkE1dnplUnhDOEVtRTQ0emN4bmZFeG1QYZCkIAAAAAAApQAAAAAAAAAG3fbh12Whk9nL4UbO63msHLSF7V9bN5E6jPWFfv8AqQ4EAQwADQEBFBIOAg8DBAUGEAcICQoFBhEBCwARCaCGAQAAAAAARAYAAAAAAAAOAwEAAAEJ"

	connection := rpc.New(os.Getenv("RPC_URL"))

	// ------------------------------------------------
	var txTemp solana.Transaction
	// txTemp.Message.UnmarshalBase64(req.TxRawMessage)
	err := txTemp.Message.UnmarshalBase64(TxRawMessage)
	if err != nil {
		logrus.Warnln("UnmarshalBase64  error", err, "TxRawMessage:", TxRawMessage)
		return
	}

	privateKey := os.Getenv("PVKEY")
	pK, err := base58.Decode(privateKey)
	if err != nil {
		logrus.Warnln("Decode  error", err)
		return
	}

	pkPrivateKey := solana.PrivateKey(pK)

	//txTemp.Message.Header.NumRequiredSignatures = 1
	//lenInst := len(txTemp.Message.Instructions)
	//txTemp.Message.Instructions[lenInst-1].Accounts[0] -= 1
	_, err = txTemp.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		return &pkPrivateKey
	})
	if err != nil {
		logrus.Warnln("Sign  error", err)
		return
	}
	req.Signatures = make([][64]byte, len(txTemp.Signatures))

	for i, sig := range txTemp.Signatures {
		req.Signatures[i] = sig
	}

	// ------------------------------------------------

	var transaction solana.Transaction
	err = transaction.Message.UnmarshalBase64(TxRawMessage)
	//transaction.Message.Header.NumRequiredSignatures = 1
	//transaction.Message.Instructions[lenInst-1].Accounts[0] -= 1

	transaction.Signatures = make([]solana.Signature, len(req.Signatures))
	for i, sig := range req.Signatures {
		transaction.Signatures[i] = solana.Signature(sig)
	}

	signature, err := connection.SendTransactionWithOpts(context.Background(), &transaction, rpc.TransactionOpts{SkipPreflight: false})

	if err != nil {
		logrus.Warnln("SendTransactionWithOpts  error", err, "transaction:", transaction)
		return
	}

	logrus.Infoln("Transaction successfully sent: ", signature)
}
