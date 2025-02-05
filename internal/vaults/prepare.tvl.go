package vaults

import (
	"strconv"

	"github.com/yearn/ydaemon/internal/strategies"
	"github.com/yearn/ydaemon/internal/utils/helpers"
	"github.com/yearn/ydaemon/internal/utils/models"
	"github.com/yearn/ydaemon/internal/utils/types/bigNumber"
)

func buildDelegated(delegatedBalanceToken *bigNumber.Int, decimals int, humanizedPrice *bigNumber.Float) float64 {
	_, delegatedTVL := helpers.FormatAmount(delegatedBalanceToken.String(), decimals)
	fHumanizedTVLPrice, _ := bigNumber.NewFloat(0).Mul(delegatedTVL, humanizedPrice).Float64()
	return fHumanizedTVLPrice
}

func prepareTVL(
	chainID uint64,
	vaultFromGraph models.TVaultFromGraph,
) float64 {
	tokenAddress := vaultFromGraph.Token.Id
	delegatedTokenAsBN := bigNumber.NewInt(0)
	fDelegatedValue := 0.0

	humanizedPrice, _ := buildTokenPrice(chainID, tokenAddress)
	fHumanizedTVLPrice := buildTVL(
		vaultFromGraph.BalanceTokens,
		int(vaultFromGraph.Token.Decimals),
		humanizedPrice,
	)

	strategiesFromMulticall := strategies.Store.StrategyMultiCallData[chainID]
	for _, strat := range vaultFromGraph.Strategies {
		multicallData, ok := strategiesFromMulticall[strat.Address]
		if !ok {
			multicallData = models.TStrategyMultiCallData{}
		}

		delegatedAssets := multicallData.DelegatedAssets
		delegatedValue := strconv.FormatFloat(
			buildDelegated(
				delegatedAssets,
				int(vaultFromGraph.Token.Decimals),
				humanizedPrice,
			), 'f', -1, 64,
		)
		stratDelegatedValueAsFloat, err := strconv.ParseFloat(delegatedValue, 64)
		if err == nil {
			delegatedTokenAsBN = delegatedTokenAsBN.Add(delegatedTokenAsBN, delegatedAssets)
			fDelegatedValue += stratDelegatedValueAsFloat
		}
	}

	return fHumanizedTVLPrice - fDelegatedValue
}
