package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	tonlib "github.com/fffilimonov/tonlib-go"
	"github.com/spf13/cobra"
	"log"
	"os"
	"strconv"
)

var sendGrammCmd = &cobra.Command{
	Use:   "sendGramm",
	Short: "Send gramm from local account to destination command",
	Long: `Send gramm command. It contains four attributes:
- path2configfile. see tonlib.config.json.example
- public key
- secret
- password
- addressDestination
- amount
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 6 {
			return fmt.Errorf("you have to use minimum six args for this commaond \n")
		}
		_, err := os.Stat(args[0])
		if err != nil {
			errors.New("please choose config path")
		}
		return nil
	},
	Run: sendGramm,
}

func sendGramm(cmd *cobra.Command, args []string) {
	confPath := args[0]
	publicKey := args[1]
	secret := args[2]
	password := args[3]
	destinationAddr := args[4]
	//// parse amount
	amount, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		log.Fatalf("failed to parse amount argument: %s as int. err: %s. ", args[5], err)
	}

	err = initClient(confPath)
	if err != nil {
		fmt.Println("init connection error: ", err)
		os.Exit(0)
	}
	pKey := tonlib.TONPrivateKey{PublicKey: publicKey, Secret: secret}

	// prepare input key
	inputKey := tonlib.InputKey{
		Type:          "inputKeyRegular",
		LocalPassword: base64.StdEncoding.EncodeToString(tonlib.SecureBytes(password)),
		Key:           pKey,
	}

	// get wallet adress info
	sourceAccState := tonlib.NewWalletInitialAccountState(pKey.PublicKey)
	senderAddr, err := tonClient.GetAccountAddress(sourceAccState, 0, 0)
	if err != nil {
		fmt.Println("get wallet address error: ", err, senderAddr)
		os.Exit(0)
	}

	state, err := tonClient.GetAccountState(*senderAddr)
	if err != nil {
		fmt.Println("unpack wallet address error: ", err)
		os.Exit(0)
	}

	fmt.Printf("Got a result: address: %v; balance :%v; last transaction id: %v. Errors: %v. \n", senderAddr.AccountAddress, state.Balance, state.LastTransactionId, err)

	// get query info
	queryInfoFee, err := tonClient.GenericCreateSendGramsQuery(
		true,
		tonlib.JSONInt64(amount),
		tonlib.NewAccountAddress(destinationAddr),
		[]byte(""),
		&inputKey,
		senderAddr,
		300, // time out of sending money not executing request
	)
	fmt.Println(fmt.Sprintf("queryInfo: %#v. err: %#v. ", queryInfoFee, err))
	if err != nil{
		fmt.Printf("Failed to create query with  error: %v \n", err)
		os.Exit(1)
	}

	// get fee
	fees, err := tonClient.QueryEstimateFees(queryInfoFee.Id, false)
	fmt.Println(fmt.Sprintf("fees: %#v. err: %#v. ", fees, err))

	// create query to send grams
	msgAction := tonlib.NewActionMsg(
		true,
		[]tonlib.MsgMessage{*tonlib.NewMsgMessage(
			tonlib.JSONInt64(amount),
			tonlib.NewMsgDataText(""),
			tonlib.NewAccountAddress(destinationAddr),
			pKey.PublicKey,
			-1,
		)},
	)
	queryInfo, err := tonClient.CreateQuery(
		msgAction,
		*senderAddr,
		sourceAccState,
		inputKey,
		300, // time out of sending money not executing request
	)
	fmt.Println(fmt.Sprintf("queryInfo: %#v. err: %#v. ", queryInfo, err))
	if err != nil {
		fmt.Printf("Failed to create query with  error: %v \n", err)
		os.Exit(1)
	}

	// send query
	ok, err := tonClient.QuerySend(queryInfo.Id)
	fmt.Println(fmt.Sprintf("send query. ok: %#v. err: %#v. ", ok, err))
}
