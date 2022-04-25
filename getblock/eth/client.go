package eth

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ybbus/jsonrpc/v3"
)

const (
	// Wei is the smallest denomination of ether.
	Wei = 1
	// KWei or "babbage" is equal to 1.000 Wei.
	KWei = 1e3
	// MWei or "lovelace" is equal to 1.000.000 Wei.
	MWei = 1e6
	// GWei or "shannon" is equal to 1.000.000.000 Wei.
	GWei = 1e9
	// TWei or "szabo" or "microether" is equal to 1.000.000.000.000 Wei.
	TWei = 1e12
	// Pwei or "finney" or "milliether" is equal to 1.000.000.000.000.000 Wei.
	Pwei = 1e15
	// Ether is the main transaction token in ethereum network. It equals to 1.000.000.000.000.000.000 Wei.
	Ether = 1e18
)

// Endpoint is default endpoint for etherium mainnet.
const Endpoint = "https://eth.getblock.io/mainnet/"

const authorizationHeaderKey = "x-api-key"

// New creates JSON-RPC client for https://eth.getblock.io/mainnet/.
func New(token string, options *ClientOptions) *Client {
	endpoint := Endpoint
	if options != nil {
		if options.Endpoint != "" {
			endpoint = options.Endpoint
		}
	}
	return &Client{
		Client: jsonrpc.NewClientWithOpts(endpoint, &jsonrpc.RPCClientOpts{
			CustomHeaders: map[string]string{authorizationHeaderKey: token},
		}),
	}
}

// Client is JSON-RPC client
type Client struct {
	Client jsonrpc.RPCClient
}

// Call sends request to JSON-RPC endpoint.
// Repeats request on 5xx error up to 5 times.
func (c *Client) Call(ctx context.Context, method string, params ...interface{}) (*jsonrpc.RPCResponse, error) {
	var r *jsonrpc.RPCResponse
	var err error
	for i := 0; i < 5; i++ {
		r, err = c.Client.Call(ctx, method, params)
		if err == nil {
			break
		}

		e, ok := err.(*jsonrpc.HTTPError)
		if ok {
			if e.Code < 500 {
				break
			}
		}
	}

	return r, err
}

// BlockNumber returns the index corresponding to the block number of the current chain head
// https://getblock.io/docs/available-nodes-methods/ETH/JSON-RPC/eth_blockNumber/.
func (c *Client) BlockNumber(ctx context.Context) (*big.Int, error) {
	r, err := c.Call(ctx, "eth_blockNumber")
	if err != nil {
		return nil, err
	}

	v, ok := r.Result.(string)
	if !ok {
		return nil, fmt.Errorf("value is not string: %v", r.Result)
	}

	return hex2int(v), nil
}

// GetBlockByNumber returns information about a block by block number
// https://getblock.io/docs/available-nodes-methods/ETH/JSON-RPC/eth_getBlockByNumber/.
func (c *Client) GetBlockByNumber(ctx context.Context, blockNumber *big.Int, detailedTransactions bool) (*Block, error) {
	bn := int2hex(blockNumber)
	r, err := c.Call(ctx, "eth_getBlockByNumber", bn, detailedTransactions)
	if err != nil {
		return nil, err
	}

	v := &Block{}
	err = r.GetObject(&v)

	return v, err

}

// Wei2ether converts wei to ether
func Wei2ether(val *big.Int) *big.Float {
	f := new(big.Float).SetInt(val)
	return f.Quo(f, big.NewFloat(Ether))
}

func hex2int(s string) *big.Int {
	i := new(big.Int)
	if s == "" {
		return i
	}

	_, ok := i.SetString(s, 0)
	if !ok {
		panic(fmt.Sprintf("valid is not big integer: %s", s))
	}

	return i
}

func int2hex(i *big.Int) string {
	return fmt.Sprintf("%#x", i)
}
