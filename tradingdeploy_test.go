package main

// Basic imports
import (
	"fmt"
	"math/big"
	"strings"

	"testing"

	"github.com/incognitochain/bridge-eth/bridge/incognito_proxy"
	"github.com/incognitochain/bridge-eth/bridge/prveth"
	"github.com/incognitochain/bridge-eth/bridge/prvbsc"
	// "github.com/incognitochain/bridge-eth/bridge/kbntrade"
	// "github.com/incognitochain/bridge-eth/bridge/uniswap"
	"github.com/incognitochain/bridge-eth/bridge/vault"
	"github.com/incognitochain/bridge-eth/bridge/vaultproxy"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TradingDeployTestSuite struct {
	*TradingTestSuite

	KyberContractAddr        common.Address
	ZRXContractAddr          common.Address
	WETHAddr                 common.Address
	UniswapRouteContractAddr common.Address
}

func NewTradingDeployTestSuite(tradingTestSuite *TradingTestSuite) *TradingDeployTestSuite {
	return &TradingDeployTestSuite{
		TradingTestSuite: tradingTestSuite,
	}
}

func (tradingDeploySuite *TradingDeployTestSuite) SetupSuite() {
	fmt.Println("Setting up the suite...")
	// 0x kovan env
	tradingDeploySuite.KyberContractAddr = common.HexToAddress("0x692f391bCc85cefCe8C237C01e1f636BbD70EA4D")
	tradingDeploySuite.ZRXContractAddr = common.HexToAddress("0xf1ec01d6236d3cd881a0bf0130ea25fe4234003e")
	tradingDeploySuite.WETHAddr = common.HexToAddress("0xd0a1e359811322d97991e03f863a0c30c2cf029c")
	// uniswap router v2
	tradingDeploySuite.UniswapRouteContractAddr = common.HexToAddress("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
}

func (tradingDeploySuite *TradingDeployTestSuite) TearDownSuite() {
	fmt.Println("Tearing down the suite...")
	tradingDeploySuite.ETHClient.Close()
}

func TestTradingDeployTestSuite(t *testing.T) {
	fmt.Println("Starting entry point...")
	tradingSuite := new(TradingTestSuite)
	suite.Run(t, tradingSuite)

	tradingDeploySuite := NewTradingDeployTestSuite(tradingSuite)
	suite.Run(t, tradingDeploySuite)
	fmt.Println("Finishing entry point...")
}

func (tradingDeploySuite *TradingDeployTestSuite) TestDeployAllContracts() {
	admin := common.HexToAddress(Admin)
	fmt.Println("Admin address:", admin.Hex())

	// Genesis committee
	// for testnet & local env
	beaconComm, bridgeComm, err := convertCommittees(
		testnetBeaconCommitteePubKeys, testnetBridgeCommitteePubKeys,
	)
	// NOTE: uncomment this block to get mainnet committees when deploying to mainnet env
	/*
		beaconComm, bridgeComm, err := convertCommittees(
			mainnetBeaconCommitteePubKeys, mainnetBridgeCommitteePubKeys,
		)
	*/

	require.Equal(tradingDeploySuite.T(), nil, err)

	// Deploy incognito_proxy
	auth := bind.NewKeyedTransactor(tradingDeploySuite.ETHPrivKey)
	auth.Value = big.NewInt(0)
	// auth.GasLimit = 4000000
	incAddr, tx, _, err := incognito_proxy.DeployIncognitoProxy(auth, tradingDeploySuite.ETHClient, admin, beaconComm, bridgeComm)
	require.Equal(tradingDeploySuite.T(), nil, err)

	// incAddr := common.HexToAddress(IncognitoProxyAddress)
	fmt.Println("deployed incognito_proxy")
	fmt.Printf("addr: %s\n", incAddr.Hex())

	// Wait until tx is confirmed
	err = wait(tradingDeploySuite.ETHClient, tx.Hash())
	require.Equal(tradingDeploySuite.T(), nil, err)

	// Deploy vault
	prevVault := common.Address{}
	vaultAddr, tx, _, err := vault.DeployVault(auth, tradingDeploySuite.ETHClient)
	require.Equal(tradingDeploySuite.T(), nil, err)
	fmt.Println("deployed vault")
	fmt.Printf("addr: %s\n", vaultAddr.Hex())

	// Wait until tx is confirmed
	err = wait(tradingDeploySuite.ETHClient, tx.Hash())
	require.Equal(tradingDeploySuite.T(), nil, err)

	vaultAbi, _ := abi.JSON(strings.NewReader(vault.VaultABI))
	input, _ := vaultAbi.Pack("initialize", prevVault)

	// Deploy vault proxy
	vaultAddr, tx, _, err = vaultproxy.DeployTransparentUpgradeableProxy(auth, tradingDeploySuite.ETHClient, vaultAddr, admin, incAddr, input)
	require.Equal(tradingDeploySuite.T(), nil, err)
	fmt.Println("deployed vault proxy")
	fmt.Printf("addr: %s\n", vaultAddr.Hex())

	err = wait(tradingDeploySuite.ETHClient, tx.Hash())
	require.Equal(tradingDeploySuite.T(), nil, err)

	// Deploy kbntrade
	// kbnTradeAddr, tx, _, err := kbntrade.DeployKBNTrade(auth, tradingDeploySuite.ETHClient, tradingDeploySuite.KyberContractAddr)
	// require.Equal(tradingDeploySuite.T(), nil, err)
	// fmt.Println("deployed kbntrade")
	// fmt.Printf("addr: %s\n", kbnTradeAddr.Hex())

	// // Wait until tx is confirmed
	// err = wait(tradingDeploySuite.ETHClient, tx.Hash())
	// require.Equal(tradingDeploySuite.T(), nil, err)

	// // Deploy uniswaptrade
	// uniswapAddr, tx, _, err := uniswap.DeployUniswapV2Trade(auth, tradingDeploySuite.ETHClient, tradingDeploySuite.UniswapRouteContractAddr)
	// require.Equal(tradingDeploySuite.T(), nil, err)
	// fmt.Println("deployed uniswap adapter")
	// fmt.Printf("addr: %s\n", uniswapAddr.Hex())

	// // Wait until tx is confirmed
	// err = wait(tradingDeploySuite.ETHClient, tx.Hash())
	// require.Equal(tradingDeploySuite.T(), nil, err)

	vaultAddr = common.HexToAddress("0x43d037a562099a4c2c95b1e2120cc43054450629")
	auth.Nonce = big.NewInt(40)
	auth.GasPrice = big.NewInt(10000000000)

	// Deploy prv erc20 token
	incAddr = common.HexToAddress("0xfAb5ab70500d1fab1e736829b4bf85e8bAC5fff2")

	prvToken, tx, _, err := prveth.DeployPrveth(auth, tradingDeploySuite.ETHClient, "Incognito", "PRV", incAddr, vaultAddr)
	require.Equal(tradingDeploySuite.T(), nil, err)
	fmt.Println("deployed prv erc20 token")
	fmt.Printf("addr: %s\n", prvToken.Hex())

	err = wait(tradingDeploySuite.ETHClient, tx.Hash())
	require.Equal(tradingDeploySuite.T(), nil, err)

	incAddr = common.HexToAddress("0xd190620159d82731F9951326Bafdc873a16CB2b1")
	auth.GasPrice = big.NewInt(10000000000)

	prvToken, tx, _, err = prvbsc.DeployPrvbsc(auth, tradingDeploySuite.BSCClient, "Incognito", "PRV", incAddr, vaultAddr)
	require.Equal(tradingDeploySuite.T(), nil, err)
	fmt.Println("deployed prv bep20 token")
	fmt.Printf("addr: %s\n", prvToken.Hex())

	err = wait(tradingDeploySuite.BSCClient, tx.Hash())
	require.Equal(tradingDeploySuite.T(), nil, err)
}

func convertCommittees(
	beaconComms []string, brigeComms []string,
) ([]common.Address, []common.Address, error) {
	beaconOld := make([]common.Address, len(beaconComms))
	for i, pk := range beaconComms {
		cpk := &CommitteePublicKey{}
		cpk.FromString(pk)
		addr, err := convertPubkeyToAddress(*cpk)
		if err != nil {
			return nil, nil, err
		}
		beaconOld[i] = addr
		fmt.Printf("beaconOld: %s\n", addr.Hex())
	}

	bridgeOld := make([]common.Address, len(brigeComms))
	for i, pk := range brigeComms {
		cpk := &CommitteePublicKey{}
		cpk.FromString(pk)
		addr, err := convertPubkeyToAddress(*cpk)
		if err != nil {
			return nil, nil, err
		}
		bridgeOld[i] = addr
		fmt.Printf("bridgeOld: %s\n", addr.Hex())
	}
	return beaconOld, bridgeOld, nil
}

func TestDisplayCommitteesMainnet(t *testing.T) {
	fmt.Println("Mainnet Committees: [")
	beaconComms := mainnetBeaconCommitteePubKeys
	brigeComms := mainnetBridgeCommitteePubKeys
	beaconOld := make([]common.Address, len(beaconComms))
	for i, pk := range beaconComms {
		cpk := &CommitteePublicKey{}
		cpk.FromString(pk)
		addr, err := convertPubkeyToAddress(*cpk)
		if err != nil {
			return
		}
		beaconOld[i] = addr
		fmt.Printf("  %s,\n", addr.Hex())
	}
	fmt.Println("]")

	bridgeOld := make([]common.Address, len(brigeComms))
	for i, pk := range brigeComms {
		cpk := &CommitteePublicKey{}
		cpk.FromString(pk)
		addr, err := convertPubkeyToAddress(*cpk)
		if err != nil {
			return
		}
		bridgeOld[i] = addr
		fmt.Printf("%s,\n", addr.Hex())
	}
}

func TestDisplayCommitteesTestnet(t *testing.T) {
	fmt.Println("Testnet Committees: [")
	beaconComms := testnetBeaconCommitteePubKeys
	brigeComms := testnetBridgeCommitteePubKeys
	beaconOld := make([]common.Address, len(beaconComms))
	for i, pk := range beaconComms {
		cpk := &CommitteePublicKey{}
		cpk.FromString(pk)
		addr, err := convertPubkeyToAddress(*cpk)
		if err != nil {
			return
		}
		beaconOld[i] = addr
		fmt.Printf("  %s,\n", addr.Hex())
	}
	fmt.Println("]")

	bridgeOld := make([]common.Address, len(brigeComms))
	for i, pk := range brigeComms {
		cpk := &CommitteePublicKey{}
		cpk.FromString(pk)
		addr, err := convertPubkeyToAddress(*cpk)
		if err != nil {
			return
		}
		bridgeOld[i] = addr
		fmt.Printf("%s,\n", addr.Hex())
	}
}
