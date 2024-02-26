package main

import (
	"crypto/ecdsa"
	"ether-go-tools/internal/blocks"
	"ether-go-tools/internal/simulation"
	"ether-go-tools/internal/utils"
	"ether-go-tools/internal/wallets"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {

	var (
		clientHttp  *ethclient.Client
		clientWs    *ethclient.Client
		err         error
		choice      int
		richPrivKey *ecdsa.PrivateKey
		richPubKey  common.Address
		config      *utils.Config
	)

	// Generate our config based on the config supplied
	// by the user in the flags
	configPath, err := utils.ParseFlags()
	utils.ErrManagement(err)

	config, err = utils.LoadConfig(configPath)
	utils.ErrManagement(err)

	richPrivKey, richPubKey, err = wallets.RetrieveKeysFromHexHashedPrivateKey(config.Connection.Rich_private_key)
	utils.ErrManagement(err)

	clientHttp, err = ethclient.Dial(config.Connection.Http_endpoint)
	utils.ErrManagement(err)

	clientWs, err = ethclient.Dial(config.Connection.Ws_endpoint)
	utils.ErrManagement(err)

	for {
		fmt.Println("Choose what do u want to do:")
		fmt.Println("1: Create a new account")
		fmt.Println("2: Retrieve information header about a block")
		fmt.Println("3: Retrieve complete information about a block")
		fmt.Println("4: Send Ethers from a rich account to an account")
		fmt.Println("5: Create Life Simulation")
		fmt.Println("6: Listening for blocks")

		fmt.Println()
		fmt.Scanf("%d", &choice)
		switch choice {
		case 1:
			fmt.Println("Create a new account")
			wallets.CreateWallet()
		case 2:
			fmt.Println("Retrieve information header about a block")
			blocks.Blockheader(clientHttp)
		case 3:
			fmt.Println("Retrieve complete information about a block")
			blocks.Blockfull(clientHttp)
		case 4:
			fmt.Println("Send Ethers from a rich account to an account")
			wallets.SendEthersCli(clientHttp, richPrivKey, richPubKey)
		case 5:
			fmt.Println("Create Life Simulation")
			simulation.Simulation(clientWs, richPrivKey, richPubKey, config.Simulation.Accounts, config.Simulation.Ethers, config.Simulation.Transactions)
		case 6:
			fmt.Println("6: Listening for blocks")
			blocks.ListeningBlock(clientWs)
		default:
			fmt.Println("Function not implemented")
		}
	}
}
