package simulation

import (
	"context"
	"crypto/ecdsa"
	"ether-go-tools/internal/utils"
	"fmt"
	"log"
	"math/big"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/crypto/sha3"
)

var (
	walletsFrom []Wallet
	walletsTo   []Wallet
	wg          sync.WaitGroup
)

type Wallet struct {
	Key        ecdsa.PrivateKey
	KeyHex     string
	Address    common.Address
	AddressHex string
}

func Simulation(wsClient *ethclient.Client, richPrivKey *ecdsa.PrivateKey, richPubKey common.Address, numWallets int, nbEthers int, numTransactions int) {
	walletsFrom = CreateWallets(numWallets)
	fmt.Println("First list of accounts")
	for _, wallet := range walletsFrom {
		fmt.Println("Public key:", wallet.AddressHex, "; Private key:", wallet.KeyHex)
	}
	SendEthers(wsClient, richPrivKey, richPubKey, walletsFrom, 1)

	walletsTo = CreateWallets(numWallets)
	fmt.Println("Second list of accounts")
	for _, wallet := range walletsTo {
		fmt.Println("Public key:", wallet.AddressHex, "; Private key:", wallet.KeyHex)
	}
	SendEthers(wsClient, richPrivKey, richPubKey, walletsTo, 1)

	time.Sleep(13 * time.Second) // Waiting for a block

	headers := make(chan *types.Header)
	sub, err := wsClient.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}
	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			block, err := wsClient.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println("Going for block number:", (block.Number().Uint64() + 1)) // 3477413
			fmt.Println(wsClient.BalanceAt(context.Background(), walletsFrom[0].Address, nil))
			SendEthersFromAPoolToAPool(wsClient, walletsFrom, walletsTo, numTransactions)
			fmt.Println(numTransactions, "transactions sended")
		}
	}

}

func NewWallets(key *ecdsa.PrivateKey, keyHex string, address common.Address, addressHex string) Wallet {
	wallet := Wallet{Key: *key, KeyHex: keyHex, Address: address, AddressHex: addressHex}
	return wallet
}

func CreateWallets(numWallets int) []Wallet {

	wallets := make([]Wallet, numWallets)

	for i := 0; i < numWallets; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatal(err)
		}
		privateKeyBytes := crypto.FromECDSA(privateKey)
		// fmt.Println("The private key:", hexutil.Encode(privateKeyBytes))

		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		}

		publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

		address := crypto.PubkeyToAddress(*publicKeyECDSA)

		hash := sha3.NewLegacyKeccak256()
		hash.Write(publicKeyBytes[1:])
		// fmt.Println("The public key:", hexutil.Encode(hash.Sum(nil)[12:]))

		wallets[i] = NewWallets(privateKey, hexutil.Encode(privateKeyBytes), address, hexutil.Encode(hash.Sum(nil)[12:]))
	}

	return wallets
}

func SendEthersFromAPoolToAPool(client *ethclient.Client, walletsFrom []Wallet, walletsTo []Wallet, numTransactions int) {

	var (
		err      error
		nbEthers int
	)
	_ = err
	nbEthers = 1
	wg.Add(1)
	go func() {
		defer wg.Done() // Ceci s'assurera que wg.Done() est appelé à la fin de la goroutine
		for i := 0; i < numTransactions/2; i++ {
			indexFrom := rand.IntN(cap(walletsFrom))
			indexTo := rand.IntN(cap(walletsTo))
			SendEthersToSpecificWallet(client, &walletsFrom[indexFrom].Key, walletsFrom[indexFrom].Address, walletsTo[indexTo], nbEthers)
			time.Sleep(time.Millisecond * 10)
		}
	}()
	func() {
		for i := 0; i < numTransactions/2; i++ {
			indexFrom := rand.IntN(cap(walletsFrom))
			indexTo := rand.IntN(cap(walletsTo))
			SendEthersToSpecificWallet(client, &walletsTo[indexTo].Key, walletsTo[indexTo].Address, walletsFrom[indexFrom], nbEthers)
			time.Sleep(time.Millisecond * 10)
		}
	}()
	wg.Wait()
}

func SendEthersToSpecificWallet(client *ethclient.Client, privateKey *ecdsa.PrivateKey, fromAddress common.Address, toWallet Wallet, nbEthers int) {
	var (
		nonce uint64
		err   error
	)
	nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
	utils.ErrManagement(err)
	// Convert nbEthers (int) en big.Int
	amount := big.NewInt(int64(nbEthers))
	// Convert Ethers to Wei (1 Ether = 1e18 Wei)
	weiValue := new(big.Int).Mul(amount, big.NewInt(0))

	value := weiValue         // in wei (1 eth)
	gasLimit := uint64(21000) // in units
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	var data []byte
	tx := types.NewTransaction(nonce, toWallet.Address, value, gasLimit, gasPrice, data)
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatal(err)
	}
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transaction")
}

func SendEthers(client *ethclient.Client, privateKey *ecdsa.PrivateKey, fromAddress common.Address, wallets []Wallet, nbEthers int) {

	var (
		nonce uint64
		err   error
	)
	i := 0
	for _, wallet := range wallets {
		nonce, err = client.PendingNonceAt(context.Background(), fromAddress)
		utils.ErrManagement(err)
		fmt.Println("Nonce:", nonce)
		// Convert nbEthers (int) en big.Int
		amount := big.NewInt(int64(nbEthers))
		// Convert Ethers to Wei (1 Ether = 1e18 Wei)
		weiValue := new(big.Int).Mul(amount, big.NewInt(1e18))

		value := weiValue         // in wei (1 eth)
		gasLimit := uint64(21000) // in units
		gasPrice, err := client.SuggestGasPrice(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		var data []byte
		tx := types.NewTransaction(nonce, wallet.Address, value, gasLimit, gasPrice, data)
		chainID, err := client.ChainID(context.Background())
		if err != nil {
			log.Fatal(err)
		}
		signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
		if err != nil {
			log.Fatal(err)
		}
		err = client.SendTransaction(context.Background(), signedTx)
		if err != nil {
			log.Fatal(err)
		}
		// fmt.Printf("tx sent: %s", signedTx.Hash().Hex())
		// fmt.Println()
		i = i + 1
	}
}
