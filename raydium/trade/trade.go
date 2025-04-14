package trade

import (
	"context"

	"github.com/gagliardetto/solana-go"
	associatedtokenaccount "github.com/gagliardetto/solana-go/programs/associated-token-account"
	computebudget "github.com/gagliardetto/solana-go/programs/compute-budget"
	"github.com/gagliardetto/solana-go/programs/system"
	"github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/katelouis/raydium-swap-go/raydium/constants"
	"github.com/katelouis/raydium-swap-go/raydium/layouts"
	"github.com/katelouis/raydium-swap-go/raydium/utils"
)

type Trade struct {
	Connection *rpc.Client
	SignerPub  *solana.PublicKey
}

type FeeConfig struct {
	MicroLamports uint64
}

func New(connection *rpc.Client, signerP *solana.PublicKey) *Trade {
	return &Trade{
		Connection: connection,
		SignerPub:  signerP,
	}
}

func (t *Trade) MakeRawSwapTx(poolKeys *layouts.ApiPoolInfoV4, amountIn *utils.TokenAmount, minAmountOut *utils.TokenAmount, feeConfig FeeConfig) (*solana.Transaction, error) {
	recent, err := t.Connection.GetLatestBlockhash(context.Background(), rpc.CommitmentFinalized)

	if err != nil {
		return &solana.Transaction{}, err
	}

	var instructions []solana.Instruction = []solana.Instruction{
		computebudget.NewSetComputeUnitLimitInstruction(600000).Build(),
		computebudget.NewSetComputeUnitPriceInstruction(feeConfig.MicroLamports).Build(),
	}

	tokenAccountIn, err := t.selectOrCreateAccount(amountIn, &instructions, "in")

	if err != nil {
		return &solana.Transaction{}, err
	}

	tokenAccountOut, err := t.selectOrCreateAccount(minAmountOut, &instructions, "out")

	if err != nil {
		return &solana.Transaction{}, err
	}
	instr, err := NewSwapV4Instruction(
		t.Connection,
		poolKeys,
		uint64(amountIn.Amount),
		uint64(minAmountOut.Amount),
		tokenAccountIn,
		tokenAccountOut,
		*t.SignerPub,
	)

	if err != nil {
		return &solana.Transaction{}, err
	}

	instructions = append(instructions, instr)

	if amountIn.Token.Mint == constants.WSOl.String() {
		closeAccInst, err := token.NewCloseAccountInstruction(
			tokenAccountIn,
			*t.SignerPub,
			*t.SignerPub,
			[]solana.PublicKey{},
		).ValidateAndBuild()

		if err != nil {
			return &solana.Transaction{}, err
		}

		instructions = append(instructions, closeAccInst)
	} else if minAmountOut.Token.Mint == constants.WSOl.String() {
		closeAccInst, err := token.NewCloseAccountInstruction(
			tokenAccountOut,
			*t.SignerPub,
			*t.SignerPub,
			[]solana.PublicKey{},
		).ValidateAndBuild()

		if err != nil {
			return &solana.Transaction{}, err
		}

		instructions = append(instructions, closeAccInst)
	}

	tx, err := solana.NewTransaction(
		instructions,
		recent.Value.Blockhash,
		solana.TransactionPayer(*t.SignerPub),
	)

	return tx, err
}

func (t *Trade) selectOrCreateAccount(amount *utils.TokenAmount, insttr *[]solana.Instruction, side string) (solana.PublicKey, error) {
	acc, err := t.Connection.GetTokenAccountsByOwner(context.Background(), *t.SignerPub, &rpc.GetTokenAccountsConfig{Mint: amount.Token.PublicKey().ToPointer()}, &rpc.GetTokenAccountsOpts{
		Encoding: "jsonParsed",
	})
	if err != nil {
		return solana.PublicKey{}, err
	}

	if len(acc.Value) > 0 {
		return acc.Value[0].Pubkey, nil
	}

	ataAddress, _, err := solana.FindAssociatedTokenAddress(*t.SignerPub, amount.Token.PublicKey())

	if err != nil {
		return solana.PublicKey{}, err
	}

	rentCost, err := t.Connection.GetMinimumBalanceForRentExemption(context.Background(), 165, rpc.CommitmentConfirmed)

	if err != nil {
		return solana.PublicKey{}, err
	}

	accountLamports := rentCost

	if amount.Mint == constants.WSOl.String() {
		if side == "in" {
			accountLamports += uint64(amount.Amount)
		}

		pubKey, seed, err := t.generatePubkeyWithSeed(*t.SignerPub, token.ProgramID)

		if err != nil {
			return solana.PublicKey{}, err
		}
		createInst, err := system.NewCreateAccountWithSeedInstruction(
			*t.SignerPub,
			seed,
			accountLamports,
			165,
			token.ProgramID,
			*t.SignerPub,
			pubKey,
			*t.SignerPub,
		).ValidateAndBuild()

		if err != nil {
			return solana.PublicKey{}, err
		}

		initInst, err := token.NewInitializeAccountInstruction(
			pubKey,
			constants.WSOl,
			*t.SignerPub,
			solana.SysVarRentPubkey,
		).ValidateAndBuild()

		if err != nil {
			return solana.PublicKey{}, err
		}

		*insttr = append(*insttr, createInst)
		*insttr = append(*insttr, initInst)

		return pubKey, nil
	}

	createAtaInst, err := associatedtokenaccount.NewCreateInstruction(
		*t.SignerPub,
		*t.SignerPub,
		amount.Token.PublicKey(),
	).ValidateAndBuild()

	if err != nil {
		return solana.PublicKey{}, err
	}
	*insttr = append(*insttr, createAtaInst)

	return ataAddress, nil
}

func (t *Trade) generatePubkeyWithSeed(from solana.PublicKey, programId solana.PublicKey) (solana.PublicKey, string, error) {
	seed := solana.NewWallet().PublicKey().String()[0:32]
	publicKey, err := solana.CreateWithSeed(from, seed, programId)

	return publicKey, seed, err
}
