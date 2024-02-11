package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Estructura de la transacción
type Transaccion struct {
	Cliente       string    `json:"cliente"`
	Monto         float64   `json:"monto"`
	Moneda        string    `json:"moenda"`
	Destino       string    `json:"destino"`
	IdTransaccion string    `json:"idTransaccion"`
	Firma         string    `json:"firma"`
	Hash          string    `json:"hash"`
	Timestamp     time.Time `json:"timestamp"`
}

// Lista de entidades sancionadas
var entidadesSancionadas = []string{"entidadSancionada1"}

// Valor promedio de transacciones del cliente
var valorPromedio float64 = 200.0

// Moneda del destino por defecto
var destinoMoneda string = "EUR"

func (s *SmartContract) CrearTransaccion(ctx contractapi.TransactionContextInterface, cliente string, monto float64, moneda string, destino string, idTransaccion string) error {
	//validaciones

	// Validar fondos
	if monto > GetSaldo(cliente) {
		fmt.Println("Fondos insuficientes")
		return nil
	}

	// Validar entidades sancionadas
	if EstaSancionado(destino) {
		fmt.Println("Entidad sancionada")
		return nil
	}

	// Validar transacción sospechosa
	if monto > valorPromedio*1.5 {
		fmt.Println("Transacción sospechosa")
		return nil
	}

	// Convertir moneda (si es necesario)
	if moneda != destinoMoneda {
		monto = ConvertirMoneda(monto, moneda, destinoMoneda)
	}

	// Crear transacción
	transaccion := Transaccion{
		Cliente:       cliente,
		Monto:         monto,
		Moneda:        moneda,
		Destino:       destino,
		IdTransaccion: idTransaccion,
		//Firma:     firma,
		//Hash:      hash,
		Timestamp: time.Now(),
	}

	transactionAsBytes, _ := json.Marshal(transaccion)

	//save in ledger
	return ctx.GetStub().PutState(idTransaccion, transactionAsBytes)
}

func (s *SmartContract) ConsultarTransaccion(ctx contractapi.TransactionContextInterface, idTransaccion string) (*Transaccion, error) {
	transactionAsBytes, err := ctx.GetStub().GetState(idTransaccion)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if transactionAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", idTransaccion)
	}

	transaccion := new(Transaccion)

	err = json.Unmarshal(transactionAsBytes, transaccion)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal error. %s", err.Error())
	}

	return transaccion, nil
}

// Función para convertir moneda
func ConvertirMoneda(monto float64, monedaOrigen string, monedaDestino string) float64 {
	var tasaCambio float64

	switch {
	case monedaOrigen == "COP" && monedaDestino == "USD":
		tasaCambio = 1.0 / 4000.0
	case monedaOrigen == "USD" && monedaDestino == "COP":
		tasaCambio = 4000.0
	case monedaOrigen == "USD" && monedaDestino == "EUR":
		tasaCambio = 0.93
	case monedaOrigen == "EUR" && monedaDestino == "USD":
		tasaCambio = 1.0 / 0.93
	case monedaOrigen == "COP" && monedaDestino == "EUR":
		tasaCambio = 1.0 / 4250.0
	case monedaOrigen == "EUR" && monedaDestino == "COP":
		tasaCambio = 4250.0
	default:
		fmt.Println("Moneda no compatible:", monedaOrigen, monedaDestino)
		return monto
	}

	return monto * tasaCambio
}

// Función para verificar si una entidad está sancionada
func EstaSancionado(entidad string) bool {
	for _, sancionado := range entidadesSancionadas {
		if strings.EqualFold(entidad, sancionado) {
			return true
		}
	}
	return false
}

// Función para obtener el saldo del cliente
func GetSaldo(cliente string) float64 {
	return 1000.0 // Saldo ficticio
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create cross-border-contract chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error create cross-border-contract chaincode: %s", err.Error())
	}
}
