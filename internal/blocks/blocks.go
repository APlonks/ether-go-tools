package blocks

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func Blockheader(client *ethclient.Client) {

	var (
		number *big.Int
		input  int
	)

	fmt.Print("Which block number ? : ")
	_, err := fmt.Scanf("%d", &input)
	if err != nil {
		fmt.Println("Erreur lors de la lecture de l'entrée :", err)
		return
	}

	number = big.NewInt(int64(input))

	header, err := client.HeaderByNumber(context.Background(), number)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("The block number :", header.Number.String()) // 5671744
	fmt.Println("The block hash", header.Hash().Hex())
	fmt.Println("The block size", header.Size())
}

func Blockfull(client *ethclient.Client) {

	var (
		number *big.Int
		input  int
	)

	fmt.Print("Which block number ? : ")
	_, err := fmt.Scanf("%d", &input)
	if err != nil {
		fmt.Println("Erreur lors de la lecture de l'entrée :", err)
		return
	}

	number = big.NewInt(int64(input))

	block, err := client.BlockByNumber(context.Background(), number)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("The block number :", block.Header().Number) // 5671744
	fmt.Println("The block hash", block.Header().Hash().Hex())
	fmt.Println("The block size", block.Header().Size())

	fmt.Println("Transactions list:", block.Transactions())

	fmt.Println()
}

func ListeningBlock(wsClient *ethclient.Client) {

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
			fmt.Println("----------")
			fmt.Println("At", time.Unix(int64(block.Time()), 0).Format("02 January 2006 15:04:05 MST,"))
			fmt.Println("block number:", block.Number().Uint64())
			fmt.Println("Hash:", block.Hash().Hex())
			fmt.Println("Transactions:", len(block.Transactions()))
			fmt.Println("Gas Limit:", block.GasLimit())
			fmt.Println("Gas Used:", block.GasUsed())
			fmt.Println("Base Fee", block.BaseFee())
			fmt.Println("----------")
			fmt.Println()
		}
	}
}
