package main

import (
	"crypto/ecdsa"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	generateCmd := flag.NewFlagSet("generate", flag.ExitOnError)
	genCount := generateCmd.Int("count", 10, "Number of addresses to generate")
	genOutput := generateCmd.String("output", "wallet_data.csv", "Output CSV file name")

	signCmd := flag.NewFlagSet("sign", flag.ExitOnError)
	signInput := signCmd.String("input", "", "Input CSV file with addresses and private keys")
	signOutput := signCmd.String("output", "signed_data.csv", "Output CSV file name")

	if len(os.Args) < 2 {
		fmt.Println("Expected 'generate' or 'sign' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "generate":
		generateCmd.Parse(os.Args[2:])
		generateWallets(*genCount, *genOutput)
	case "sign":
		signCmd.Parse(os.Args[2:])
		if *signInput == "" {
			log.Fatal("Input file is required for sign command")
		}
		signMessages(*signInput, *signOutput)
	default:
		fmt.Println("Expected 'generate' or 'sign' subcommands")
		os.Exit(1)
	}
}

func generateWallets(count int, outputFile string) {
	file, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Cannot create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"Address", "PrivateKey"})

	for i := 0; i < count; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("Failed to generate key: %v", err)
		}

		privateKeyBytes := crypto.FromECDSA(privateKey)
		privateKeyString := hexutil.Encode(privateKeyBytes)[2:]

		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

		writer.Write([]string{address, privateKeyString})
	}

	fmt.Printf("Generated %d wallet entries and saved to %s\n", count, outputFile)
}

func signMessages(inputFile, outputFile string) {
	inFile, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("Cannot open input file: %v", err)
	}
	defer inFile.Close()

	outFile, err := os.Create(outputFile)
	if err != nil {
		log.Fatalf("Cannot create output file: %v", err)
	}
	defer outFile.Close()

	reader := csv.NewReader(inFile)
	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	writer.Write([]string{"walletAddr", "type", "text", "signature"})

	// Skip header
	_, err = reader.Read()
	if err != nil {
		log.Fatalf("Failed to read CSV header: %v", err)
	}

	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file
		}

		address := record[0]
		privateKeyHex := record[1]

		privateKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			log.Fatalf("Failed to parse private key: %v", err)
		}

		message := generateMessage()
		signature, err := signMessage(message, privateKey)
		if err != nil {
			log.Fatalf("Failed to sign message: %v", err)
		}

		writer.Write([]string{address, "wallet", message, signature})
	}

	fmt.Printf("Generated signed messages and saved to %s\n", outputFile)
}

func generateMessage() string {
	uniqueNumber := rand.Intn(1000000000)
	expirationTime := time.Now().UTC().Add(2 * time.Minute).Format(time.RFC3339)

	return fmt.Sprintf("Hello! airdrop.carv.io asks you to sign this message to confirm your ownership of the address. This action will not cost any gas fee. \n\nHere is a unique number: %d\n\nExpiration time: %s\n", uniqueNumber, expirationTime)
}

func signMessage(message string, privateKey *ecdsa.PrivateKey) (string, error) {
	fullMessage := fmt.Sprintf("\x19Ethereum Signed Message:\n%d%s", len(message), message)
	hash := crypto.Keccak256Hash([]byte(fullMessage))
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", err
	}

	signature[64] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper

	return hexutil.Encode(signature), nil
}
