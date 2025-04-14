package raydium

import (
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/katelouis/raydium-swap-go/raydium/liquidity"
	"github.com/katelouis/raydium-swap-go/raydium/pool"
	"github.com/katelouis/raydium-swap-go/raydium/trade"
)

type Raydium struct {
	connection *rpc.Client
	Pool       *pool.Pool
	Liquidity  *liquidity.Liquidity
	SignerPub  solana.PublicKey
	Trade      *trade.Trade
}

func New(connection *rpc.Client, pubKey string) *Raydium {
	publicKey := solana.MustPublicKeyFromBase58(pubKey)
	r := &Raydium{
		connection: connection,
		Pool:       pool.New(connection),
		Liquidity:  liquidity.New(connection),
		SignerPub:  publicKey,
		Trade:      trade.New(connection, &publicKey),
	}

	return r
}
