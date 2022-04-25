package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"sort"

	"github.com/ofen/wtop/getblock/eth"
	"golang.org/x/sync/errgroup"
)

var (
	n = flag.Int("n", 100, "")
	t = flag.String("t", "", "")
	v = flag.Bool("v", false, "")
)

var usage = `Usage: wtop [options...] [<block_number>]
Options:
  -n  Number of look-behind blocks. Should be between 1 and 500. Default is 100.
  -t  API token (see https://getblock.io/docs/get-started/auth-with-api-key/).
  -v  Verbose output.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}

	flag.Parse()

	blockNumber := new(big.Int)

	if flag.NArg() > 0 {
		if _, ok := blockNumber.SetString(flag.Arg(0), 10); !ok {
			usageAndExit("block number should be decimal integer")
		}

		if blockNumber.Cmp(big.NewInt(0)) < 0 {
			usageAndExit("block number should be positive decimal integer")
		}
	}

	if *t == "" {
		usageAndExit("api token required")
	}

	if *n < 1 {
		usageAndExit("-n is too low")
	}

	if *n > 500 {
		usageAndExit("-n is too high")
	}

	ctx := context.Background()
	client := eth.New(*t, nil)
	head, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Use head if block number not provided
	if flag.NArg() < 1 {
		blockNumber = new(big.Int).Set(head)
	}

	if blockNumber.Cmp(head) > 0 {
		usageAndExit(fmt.Sprintf("block number is out of range: %d", blockNumber))
	}

	wallets, err := listWallets(ctx, client, *n, blockNumber)
	if err != nil {
		log.Fatal(err)
	}

	lessBlockNumber := new(big.Int).Set(blockNumber)
	lessBlockNumber.Sub(lessBlockNumber, big.NewInt(int64(*n)))

	log.Printf("blocks %v..%v\n", lessBlockNumber, blockNumber)

	if len(wallets) < 1 {
		log.Println("there are no transactions")
		return
	}

	sortWallets(wallets)

	min := wallets[0]
	max := wallets[len(wallets)-1]

	log.Printf("%+f ETH %s\n", eth.Wei2ether(max.amount), max.address)
	log.Printf("%+f ETH %s\n", eth.Wei2ether(min.amount), min.address)

}

func listWallets(ctx context.Context, client *eth.Client, numberOfBlocks int, blockNumber *big.Int) ([]*wallet, error) {
	g, ctx := errgroup.WithContext(ctx)
	bn := new(big.Int).Set(blockNumber)

	ch := make(chan eth.Transaction)
	for i := 0; i < numberOfBlocks; i++ {
		g.Go(func() error {
			b, err := client.GetBlockByNumber(ctx, bn, true)
			if err != nil {
				return err
			}

			for _, t := range b.Transactions {
				if *v {
					log.Printf("%f ETH %s -> %s", eth.Wei2ether(t.Value), t.From, t.To)
				}
				ch <- t
			}

			return nil
		})

		bn.Sub(bn, big.NewInt(1))
		// stop if it's last block
		if bn.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	go func() {
		g.Wait()
		close(ch)
	}()

	m := make(map[string]*big.Int)
	for t := range ch {
		v, ok := m[t.From]
		if !ok {
			v = new(big.Int)
		}

		m[t.From] = v.Sub(v, t.Value)

		v, ok = m[t.To]
		if !ok {
			v = new(big.Int)
		}

		m[t.To] = v.Add(v, t.Value)
	}

	wallets := make([]*wallet, 0, len(m))
	for k, v := range m {
		wallets = append(wallets, &wallet{address: k, amount: v})
	}

	return wallets, g.Wait()
}

type wallet struct {
	address string
	amount  *big.Int
}

func sortWallets(x []*wallet) {
	sort.Slice(x, func(i, j int) bool {
		if r := x[i].amount.Cmp(x[j].amount); r < 0 {
			return true
		}
		return false
	})
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(2)
}
